package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	goRuntime "runtime"
	"strings"
	"sync"
	"syscall"
	"time"

	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"

	"github.com/gorilla/websocket"
)

// ==================== Data Types ====================

type GatewayStatus struct {
	Running    bool      `json:"running"`
	PID        int       `json:"pid"`
	Port       int       `json:"port"`
	StartedAt  time.Time `json:"startedAt"`
	ConfigPath string    `json:"configPath"`
	Uptime     string    `json:"uptime"`
}

type ChatMessage struct {
	ID        string    `json:"id"`
	Role      string    `json:"role"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
	Session   string    `json:"session"`
	Channel   string    `json:"channel"`
	Streaming bool      `json:"streaming"`
}

type SessionInfo struct {
	Key       string `json:"key"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	Path      string `json:"path"`
}

// ChannelMessage represents a real-time channel message from gateway
type ChannelMessage struct {
	Channel  string `json:"channel"`
	ChatID   string `json:"chat_id"`
	SenderID string `json:"sender_id"`
	Content  string `json:"content"`
	Role     string `json:"role"`
	Time     string `json:"time"`
}

// Notification represents a desktop notification
type Notification struct {
	Title   string `json:"title"`
	Body    string `json:"body"`
	Icon    string `json:"icon,omitempty"`
	Channel string `json:"channel,omitempty"`
}

// ==================== App ====================

type App struct {
	ctx        context.Context
	configPath string
	configMu   sync.RWMutex

	// Gateway process
	gateway     *exec.Cmd
	gatewayMu   sync.Mutex
	gwStatus    GatewayStatus
	gwCancelFn  context.CancelFunc

	// Gateway WebSocket bridge
	wsConn    *websocket.Conn
	wsMu      sync.Mutex
	wsRunning bool

	// Chat
	messages  []ChatMessage
	msgMu     sync.RWMutex
	msgCounter int64

	// Sessions
	sessions  []SessionInfo
	sessionMu sync.RWMutex

	// Channel messages (real-time feed)
	channelMsgs  []ChannelMessage
	channelMsgMu sync.RWMutex

	// Gateway logs
	gwLogs   []string
	gwLogsMu sync.RWMutex

	// Window state
	hidden     bool
	hiddenMu   sync.Mutex
}

func NewApp() *App {
	home, _ := os.UserHomeDir()
	configPath := filepath.Join(home, ".nanobot", "config.json")

	return &App{
		configPath:   configPath,
		messages:     make([]ChatMessage, 0),
		sessions:     make([]SessionInfo, 0),
		channelMsgs:  make([]ChannelMessage, 0),
		gwLogs:       make([]string, 0),
	}
}

// ==================== Lifecycle ====================

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	go a.pollSessions()
	go a.autoStartGateway()
	go a.watchSignals()
}

func (a *App) domReady(ctx context.Context) {
	wailsRuntime.EventsEmit(a.ctx, "gateway:status", a.GetGatewayStatus())
}

func (a *App) shutdown(ctx context.Context) {
	a.StopGateway()
	a.stopWSBridge()
}

func (a *App) beforeClose(ctx context.Context) (prevent bool) {
	// Hide to tray instead of closing
	a.HideWindow()
	return true // prevent close
}

// watchSignals handles OS signals
func (a *App) watchSignals() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)
	<-sigChan
	a.StopGateway()
	a.stopWSBridge()
	os.Exit(0)
}

// autoStartGateway checks if gateway should auto-start
func (a *App) autoStartGateway() {
	time.Sleep(1 * time.Second)
	// Check if any channels are enabled
	data, err := os.ReadFile(a.configPath)
	if err != nil {
		return
	}
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return
	}
	channels, ok := raw["channels"].(map[string]interface{})
	if !ok {
		return
	}
	for _, ch := range channels {
		if cfg, ok := ch.(map[string]interface{}); ok {
			if enabled, ok := cfg["enabled"].(bool); ok && enabled {
				// Auto-start gateway if any channel is enabled
				err := a.StartGateway()
				if err == nil {
					a.addLog("[auto] Gateway auto-started (enabled channels detected)")
				}
				return
			}
		}
	}
}

// ==================== Window Management ====================

func (a *App) NavigateTo(page string) {
	wailsRuntime.EventsEmit(a.ctx, "navigate", page)
}

func (a *App) Quit() {
	a.StopGateway()
	a.stopWSBridge()
	wailsRuntime.Quit(a.ctx)
}

func (a *App) HideWindow() {
	a.hiddenMu.Lock()
	a.hidden = true
	a.hiddenMu.Unlock()
	wailsRuntime.WindowHide(a.ctx)
}

func (a *App) ShowWindow() {
	a.hiddenMu.Lock()
	a.hidden = false
	a.hiddenMu.Unlock()
	wailsRuntime.WindowShow(a.ctx)
	wailsRuntime.WindowUnminimise(a.ctx)
}

func (a *App) ToggleFullscreen() {
	wailsRuntime.WindowToggleMaximise(a.ctx)
}

func (a *App) IsHidden() bool {
	a.hiddenMu.Lock()
	defer a.hiddenMu.Unlock()
	return a.hidden
}

// ShowAndNavigate shows window and navigates
func (a *App) ShowAndNavigate(page string) {
	a.ShowWindow()
	a.NavigateTo(page)
}

// ==================== Config Management ====================

func (a *App) GetConfig() string {
	a.configMu.RLock()
	defer a.configMu.RUnlock()
	data, err := os.ReadFile(a.configPath)
	if err != nil {
		return "{}"
	}
	return string(data)
}

func (a *App) SaveConfig(configJSON string) error {
	var test map[string]interface{}
	if err := json.Unmarshal([]byte(configJSON), &test); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}
	a.configMu.Lock()
	defer a.configMu.Unlock()

	backup := a.configPath + ".bak"
	if data, err := os.ReadFile(a.configPath); err == nil {
		os.WriteFile(backup, data, 0600)
	}
	formatted, err := formatJSON(configJSON)
	if err != nil {
		return fmt.Errorf("failed to format JSON: %w", err)
	}
	if err := os.WriteFile(a.configPath, []byte(formatted), 0600); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}
	wailsRuntime.EventsEmit(a.ctx, "config:saved", true)
	return nil
}

// ==================== Provider Management ====================

func (a *App) SetProviderAPIKey(provider string, apiKey string) error {
	data, err := os.ReadFile(a.configPath)
	if err != nil {
		return err
	}
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	providers, ok := raw["providers"].(map[string]interface{})
	if !ok {
		providers = make(map[string]interface{})
		raw["providers"] = providers
	}
	p, ok := providers[provider].(map[string]interface{})
	if !ok {
		p = make(map[string]interface{})
	}
	p["apiKey"] = apiKey
	providers[provider] = p
	newData, _ := json.MarshalIndent(raw, "", "  ")
	if err := os.WriteFile(a.configPath, newData, 0600); err != nil {
		return err
	}
	wailsRuntime.EventsEmit(a.ctx, "config:saved", true)
	return nil
}

func (a *App) GetProviders() map[string]interface{} {
	data, err := os.ReadFile(a.configPath)
	if err != nil {
		return map[string]interface{}{}
	}
	var raw map[string]interface{}
	json.Unmarshal(data, &raw)
	if providers, ok := raw["providers"].(map[string]interface{}); ok {
		return providers
	}
	return map[string]interface{}{}
}

// ==================== Channel Management ====================

func (a *App) GetChannels() map[string]interface{} {
	data, err := os.ReadFile(a.configPath)
	if err != nil {
		return map[string]interface{}{}
	}
	var raw map[string]interface{}
	json.Unmarshal(data, &raw)
	if channels, ok := raw["channels"].(map[string]interface{}); ok {
		return channels
	}
	return map[string]interface{}{}
}

func (a *App) SetChannelField(channel string, field string, value string) error {
	data, err := os.ReadFile(a.configPath)
	if err != nil {
		return err
	}
	var raw map[string]interface{}
	json.Unmarshal(data, &raw)
	channels, ok := raw["channels"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("channels not found")
	}
	ch, ok := channels[channel].(map[string]interface{})
	if !ok {
		return fmt.Errorf("channel %s not found", channel)
	}
	var parsed interface{}
	if err := json.Unmarshal([]byte(value), &parsed); err != nil {
		parsed = value
	}
	ch[field] = parsed
	newData, _ := json.MarshalIndent(raw, "", "  ")
	if err := os.WriteFile(a.configPath, newData, 0600); err != nil {
		return err
	}
	wailsRuntime.EventsEmit(a.ctx, "config:saved", true)
	return nil
}

// ==================== Gateway Management ====================

func (a *App) StartGateway() error {
	a.gatewayMu.Lock()
	defer a.gatewayMu.Unlock()

	if a.gwStatus.Running {
		return fmt.Errorf("gateway already running (PID %d)", a.gwStatus.PID)
	}

	nanobotBin, err := findNanobot()
	if err != nil {
		return fmt.Errorf("nanobot not found: %w\nInstall: uv tool install nanobot-ai", err)
	}

	port := 18790
	if data, err := os.ReadFile(a.configPath); err == nil {
		var raw map[string]interface{}
		json.Unmarshal(data, &raw)
		if gw, ok := raw["gateway"].(map[string]interface{}); ok {
			if p, ok := gw["port"].(float64); ok {
				port = int(p)
			}
		}
	}

	args := []string{"gateway", "--port", fmt.Sprintf("%d", port), "--config", a.configPath}
	ctx, cancel := context.WithCancel(context.Background())

	cmd := exec.CommandContext(ctx, nanobotBin, args...)
	cmd.Stdout = NewLineWriter(a.ctx, "gateway:stdout", func(line string) {
		a.addLog(line)
	})
	cmd.Stderr = NewLineWriter(a.ctx, "gateway:stderr", func(line string) {
		a.addLog("[err] " + line)
	})
	cmd.Env = os.Environ()

	if err := cmd.Start(); err != nil {
		cancel()
		return fmt.Errorf("failed to start gateway: %w", err)
	}

	a.gateway = cmd
	a.gwCancelFn = cancel
	a.gwStatus = GatewayStatus{
		Running:    true,
		PID:        cmd.Process.Pid,
		Port:       port,
		StartedAt:  time.Now(),
		ConfigPath: a.configPath,
	}

	a.emitGatewayStatus()
	wailsRuntime.EventsEmit(a.ctx, "gateway:started", true)
	a.addLog(fmt.Sprintf("[gateway] Started on port %d (PID %d)", port, cmd.Process.Pid))

	go func() {
		err := cmd.Wait()
		a.gatewayMu.Lock()
		a.gwStatus = GatewayStatus{}
		a.gwCancelFn = nil
		a.gatewayMu.Unlock()

		if err != nil && ctx.Err() == nil {
			a.addLog(fmt.Sprintf("[gateway] Stopped with error: %v", err))
		} else {
			a.addLog("[gateway] Stopped")
		}
		a.stopWSBridge()
		wailsRuntime.EventsEmit(a.ctx, "gateway:status", a.GetGatewayStatus())
	}()

	// Connect WebSocket bridge after short delay
	go func() {
		time.Sleep(2 * time.Second)
		a.startWSBridge(port)
	}()

	return nil
}

func (a *App) StopGateway() error {
	a.gatewayMu.Lock()
	defer a.gatewayMu.Unlock()

	if a.gateway == nil || a.gateway.Process == nil {
		return nil
	}

	a.addLog("[gateway] Stopping...")
	a.stopWSBridge()

	if a.gwCancelFn != nil {
		a.gwCancelFn()
	}
	// Give it a moment for graceful shutdown
	done := make(chan error, 1)
	go func() {
		done <- a.gateway.Wait()
	}()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		a.gateway.Process.Kill()
	}

	a.gwStatus = GatewayStatus{}
	wailsRuntime.EventsEmit(a.ctx, "gateway:status", GatewayStatus{})
	return nil
}

func (a *App) RestartGateway() error {
	if err := a.StopGateway(); err != nil {
		return err
	}
	time.Sleep(500 * time.Millisecond)
	return a.StartGateway()
}

func (a *App) GetGatewayStatus() GatewayStatus {
	a.gatewayMu.Lock()
	s := a.gwStatus
	a.gatewayMu.Unlock()

	if s.Running {
		s.Uptime = formatDuration(time.Since(s.StartedAt))
	}
	return s
}

func (a *App) emitGatewayStatus() {
	wailsRuntime.EventsEmit(a.ctx, "gateway:status", a.GetGatewayStatus())
}

// ==================== Gateway Logs ====================

func (a *App) GetGatewayLogs() string {
	a.gwLogsMu.RLock()
	defer a.gwLogsMu.RUnlock()
	return strings.Join(a.gwLogs, "\n")
}

func (a *App) ClearGatewayLogs() {
	a.gwLogsMu.Lock()
	defer a.gwLogsMu.Unlock()
	a.gwLogs = make([]string, 0)
	wailsRuntime.EventsEmit(a.ctx, "gateway:logs:cleared", true)
}

func (a *App) addLog(line string) {
	a.gwLogsMu.Lock()
	a.gwLogs = append(a.gwLogs, time.Now().Format("15:04:05")+" "+line)
	if len(a.gwLogs) > 2000 {
		a.gwLogs = a.gwLogs[len(a.gwLogs)-1000:]
	}
	a.gwLogsMu.Unlock()
}

// ==================== Gateway WebSocket Bridge ====================

// startWSBridge connects to nanobot gateway's internal event stream
// This bridges channel messages to the desktop UI in real-time
func (a *App) startWSBridge(port int) {
	a.wsMu.Lock()
	if a.wsRunning {
		a.wsMu.Unlock()
		return
	}
	a.wsRunning = true
	a.wsMu.Unlock()

	defer func() {
		a.wsMu.Lock()
		a.wsRunning = false
		a.wsMu.Unlock()
	}()

	// Try to connect to gateway's WebSocket
	gwURL := fmt.Sprintf("ws://127.0.0.1:%d/ws/desktop", port)
	u, _ := url.Parse(gwURL)

	for i := 0; i < 10; i++ {
		if !a.gwStatus.Running {
			return
		}
		conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
		if err == nil {
			a.wsMu.Lock()
			a.wsConn = conn
			a.wsMu.Unlock()
			a.addLog("[ws] Connected to gateway WebSocket")
			a.readWSLoop(conn)
			return
		}
		time.Sleep(2 * time.Second)
	}

	// If native WS fails, fall back to HTTP polling
	a.addLog("[ws] Native WS unavailable, using HTTP polling fallback")
	a.httpPollLoop(port)
}

func (a *App) readWSLoop(conn *websocket.Conn) {
	defer conn.Close()
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			a.addLog("[ws] Disconnected: " + err.Error())
			return
		}
		// Parse and emit
		var data map[string]interface{}
		if json.Unmarshal(msg, &data) == nil {
			eventType, _ := data["type"].(string)
			switch eventType {
			case "channel:message":
				chMsg := ChannelMessage{
					Channel:  strVal(data["channel"]),
					ChatID:   strVal(data["chat_id"]),
					SenderID: strVal(data["sender_id"]),
					Content:  strVal(data["content"]),
					Role:     strVal(data["role"]),
					Time:     strVal(data["time"]),
				}
				a.channelMsgMu.Lock()
				a.channelMsgs = append(a.channelMsgs, chMsg)
				if len(a.channelMsgs) > 500 {
					a.channelMsgs = a.channelMsgs[len(a.channelMsgs)-200:]
				}
				a.channelMsgMu.Unlock()
				wailsRuntime.EventsEmit(a.ctx, "channel:message", chMsg)
			case "agent:response":
				wailsRuntime.EventsEmit(a.ctx, "agent:response", data)
			default:
				wailsRuntime.EventsEmit(a.ctx, "gateway:event", data)
			}
		}
	}
}

// httpPollLoop is a fallback that polls gateway HTTP endpoints
func (a *App) httpPollLoop(port int) {
	baseURL := fmt.Sprintf("http://127.0.0.1:%d", port)
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		if !a.gwStatus.Running {
			return
		}
		// Health check
		resp, err := http.Get(baseURL + "/health")
		if err != nil {
			continue
		}
		resp.Body.Close()
		if resp.StatusCode == 200 {
			wailsRuntime.EventsEmit(a.ctx, "gateway:health", map[string]bool{"ok": true})
		}
	}
}

func (a *App) stopWSBridge() {
	a.wsMu.Lock()
	defer a.wsMu.Unlock()
	if a.wsConn != nil {
		a.wsConn.Close()
		a.wsConn = nil
	}
	a.wsRunning = false
}

// ==================== Channel Messages ====================

func (a *App) GetChannelMessages() []ChannelMessage {
	a.channelMsgMu.RLock()
	defer a.channelMsgMu.RUnlock()
	result := make([]ChannelMessage, len(a.channelMsgs))
	copy(result, a.channelMsgs)
	return result
}

func (a *App) ClearChannelMessages() {
	a.channelMsgMu.Lock()
	defer a.channelMsgMu.Unlock()
	a.channelMsgs = make([]ChannelMessage, 0)
	wailsRuntime.EventsEmit(a.ctx, "channel:messages:cleared", true)
}

// ==================== Agent Chat ====================

func (a *App) SendMessage(message string) error {
	if message == "" {
		return fmt.Errorf("empty message")
	}

	nanobotBin, err := findNanobot()
	if err != nil {
		return err
	}

	a.msgCounter++
	msgID := fmt.Sprintf("msg_%d_%d", time.Now().UnixMilli(), a.msgCounter)

	userMsg := ChatMessage{
		ID:        msgID,
		Role:      "user",
		Content:   message,
		Timestamp: time.Now(),
		Session:   "desktop",
		Channel:   "desktop",
	}
	a.msgMu.Lock()
	a.messages = append(a.messages, userMsg)
	a.msgMu.Unlock()
	wailsRuntime.EventsEmit(a.ctx, "chat:message", userMsg)

	// Create placeholder for assistant response
	respID := fmt.Sprintf("msg_%d_r_%d", time.Now().UnixMilli(), a.msgCounter)
	respMsg := ChatMessage{
		ID:        respID,
		Role:      "assistant",
		Content:   "",
		Timestamp: time.Now(),
		Session:   "desktop",
		Channel:   "desktop",
		Streaming: true,
	}
	a.msgMu.Lock()
	a.messages = append(a.messages, respMsg)
	a.msgMu.Unlock()
	wailsRuntime.EventsEmit(a.ctx, "chat:message", respMsg)

	go func() {
		args := []string{"agent", "--message", message, "--config", a.configPath}
		cmd := exec.Command(nanobotBin, args...)

		// Stream output via pipe
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			a.finishStream(respID, fmt.Sprintf("Error creating pipe: %v", err))
			return
		}
		cmd.Stderr = cmd.Stdout

		if err := cmd.Start(); err != nil {
			a.finishStream(respID, fmt.Sprintf("Error: %v", err))
			return
		}

		// Read streaming output
		scanner := bufio.NewScanner(stdout)
		scanner.Buffer(make([]byte, 64*1024), 1024*1024)
		var fullContent strings.Builder
		for scanner.Scan() {
			line := scanner.Text()
			fullContent.WriteString(line)
			fullContent.WriteString("\n")

			// Update streaming message
			a.msgMu.Lock()
			for i := range a.messages {
				if a.messages[i].ID == respID {
					a.messages[i].Content = fullContent.String()
				}
			}
			a.msgMu.Unlock()
			wailsRuntime.EventsEmit(a.ctx, "chat:stream", map[string]string{
				"id":      respID,
				"content": fullContent.String(),
			})
		}

		cmd.Wait()
		a.finishStream(respID, fullContent.String())
	}()

	return nil
}

func (a *App) finishStream(msgID string, content string) {
	a.msgMu.Lock()
	for i := range a.messages {
		if a.messages[i].ID == msgID {
			a.messages[i].Content = strings.TrimSpace(content)
			a.messages[i].Streaming = false
		}
	}
	a.msgMu.Unlock()
	wailsRuntime.EventsEmit(a.ctx, "chat:stream:done", map[string]string{"id": msgID})
}

func (a *App) GetMessages() []ChatMessage {
	a.msgMu.RLock()
	defer a.msgMu.RUnlock()
	result := make([]ChatMessage, len(a.messages))
	copy(result, a.messages)
	return result
}

func (a *App) ClearMessages() {
	a.msgMu.Lock()
	defer a.msgMu.Unlock()
	a.messages = make([]ChatMessage, 0)
	wailsRuntime.EventsEmit(a.ctx, "chat:cleared", true)
}

// ==================== Sessions ====================

func (a *App) GetSessions() []SessionInfo {
	a.sessionMu.RLock()
	defer a.sessionMu.RUnlock()
	result := make([]SessionInfo, len(a.sessions))
	copy(result, a.sessions)
	return result
}

func (a *App) pollSessions() {
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()
	a.scanSessions()
	for range ticker.C {
		a.scanSessions()
	}
}

func (a *App) scanSessions() {
	home, _ := os.UserHomeDir()
	sessionsDir := filepath.Join(home, ".nanobot", "workspace", "sessions")

	a.sessionMu.Lock()
	defer a.sessionMu.Unlock()

	a.sessions = make([]SessionInfo, 0)
	filepath.WalkDir(sessionsDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".jsonl") {
			return nil
		}
		f, err := os.Open(path)
		if err != nil {
			return nil
		}
		defer f.Close()

		scanner := bufio.NewScanner(f)
		scanner.Buffer(make([]byte, 4096), 4096)
		if scanner.Scan() {
			var meta map[string]interface{}
			if json.Unmarshal([]byte(scanner.Text()), &meta) == nil {
				if meta["_type"] == "metadata" {
					a.sessions = append(a.sessions, SessionInfo{
						Key:       strVal(meta["key"]),
						CreatedAt: strVal(meta["created_at"]),
						UpdatedAt: strVal(meta["updated_at"]),
						Path:      path,
					})
				}
			}
		}
		return nil
	})

	wailsRuntime.EventsEmit(a.ctx, "sessions:updated", a.sessions)
}

// ==================== System Info ====================

func (a *App) GetSystemInfo() map[string]string {
	nanobotBin, _ := findNanobot()
	info := map[string]string{
		"os":               goRuntime.GOOS,
		"arch":             goRuntime.GOARCH,
		"nanobot":          nanobotBin,
		"configPath":       a.configPath,
		"goVersion":        goRuntime.Version(),
		"version":          "1.1.0",
		"channelMessages":  fmt.Sprintf("%d", len(a.channelMsgs)),
		"sessions":         fmt.Sprintf("%d", len(a.sessions)),
	}

	// Count enabled channels
	data, err := os.ReadFile(a.configPath)
	if err == nil {
		var raw map[string]interface{}
		json.Unmarshal(data, &raw)
		enabled := 0
		if channels, ok := raw["channels"].(map[string]interface{}); ok {
			for _, ch := range channels {
				if cfg, ok := ch.(map[string]interface{}); ok {
					if e, ok := cfg["enabled"].(bool); ok && e {
						enabled++
					}
				}
			}
		}
		info["enabledChannels"] = fmt.Sprintf("%d", enabled)
	}
	return info
}

func (a *App) OpenInFinder(path string) {
	switch goRuntime.GOOS {
	case "darwin":
		exec.Command("open", "-R", path).Start()
	case "windows":
		exec.Command("explorer", "/select,", path).Start()
	default:
		exec.Command("xdg-open", filepath.Dir(path)).Start()
	}
}

func (a *App) OpenURL(rawurl string) {
	switch goRuntime.GOOS {
	case "darwin":
		exec.Command("open", rawurl).Start()
	case "windows":
		exec.Command("rundll32", "url.dll,FileProtocolHandler", rawurl).Start()
	default:
		exec.Command("xdg-open", rawurl).Start()
	}
}

// ==================== Helpers ====================

func findNanobot() (string, error) {
	if path, err := exec.LookPath("nanobot"); err == nil {
		return path, nil
	}
	home, _ := os.UserHomeDir()
	candidates := []string{
		filepath.Join(home, ".local", "bin", "nanobot"),
	}
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return c, nil
		}
	}
	return "", fmt.Errorf("nanobot not found in PATH or ~/.local/bin")
}

func strVal(v interface{}) string {
	if v == nil {
		return ""
	}
	switch val := v.(type) {
	case string:
		return val
	case float64:
		return fmt.Sprintf("%.0f", val)
	default:
		return fmt.Sprintf("%v", val)
	}
}

func formatJSON(input string) (string, error) {
	var obj interface{}
	if err := json.Unmarshal([]byte(input), &obj); err != nil {
		return "", err
	}
	out, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func formatDuration(d time.Duration) string {
	d = d.Round(time.Second)
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	s := int(d.Seconds()) % 60
	if h > 0 {
		return fmt.Sprintf("%dh %dm %ds", h, m, s)
	}
	if m > 0 {
		return fmt.Sprintf("%dm %ds", m, s)
	}
	return fmt.Sprintf("%ds", s)
}

// ==================== Event Writers ====================

// LineWriter writes output line by line with event emission
type LineWriter struct {
	ctx     context.Context
	event   string
	onLine  func(string)
	buf     []byte
	bufMu   sync.Mutex
}

func NewLineWriter(ctx context.Context, event string, onLine func(string)) *LineWriter {
	return &LineWriter{
		ctx:    ctx,
		event:  event,
		onLine: onLine,
		buf:    make([]byte, 0, 4096),
	}
}

func (w *LineWriter) Write(p []byte) (n int, err error) {
	n = len(p)
	w.bufMu.Lock()
	w.buf = append(w.buf, p...)
	// Emit raw
	wailsRuntime.EventsEmit(w.ctx, w.event, string(p))
	// Check for complete lines
	for {
		idx := bytesIndexByte(w.buf, '\n')
		if idx < 0 {
			break
		}
		line := string(w.buf[:idx])
		w.buf = w.buf[idx+1:]
		if w.onLine != nil {
			w.onLine(strings.TrimRight(line, "\r"))
		}
	}
	w.bufMu.Unlock()
	return n, nil
}

func bytesIndexByte(s []byte, c byte) int {
	for i, b := range s {
		if b == c {
			return i
		}
	}
	return -1
}

var _ io.Writer = (*LineWriter)(nil)
