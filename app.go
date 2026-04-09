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
	BotID      string    `json:"botId"`
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

type ChannelMessage struct {
	Channel  string `json:"channel"`
	ChatID   string `json:"chat_id"`
	SenderID string `json:"sender_id"`
	Content  string `json:"content"`
	Role     string `json:"role"`
	Time     string `json:"time"`
}

// ==================== App ====================

type App struct {
	ctx context.Context
	botMgr *BotManager

	// Active bot's gateway process
	gateway    *exec.Cmd
	gatewayMu  sync.Mutex
	gwStatus   GatewayStatus
	gwCancelFn context.CancelFunc

	// Gateway WebSocket bridge
	wsConn    *websocket.Conn
	wsMu      sync.Mutex
	wsRunning bool

	// Chat
	messages   []ChatMessage
	msgMu      sync.RWMutex
	msgCounter int64

	// Sessions
	sessions  []SessionInfo
	sessionMu sync.RWMutex

	// Channel messages (real-time feed)
	channelMsgs  []ChannelMessage
	channelMsgMu sync.RWMutex

	// Gateway logs (per active bot)
	gwLogs   []string
	gwLogsMu sync.RWMutex

	// Window state
	hidden   bool
	hiddenMu sync.Mutex
}

func NewApp() *App {
	return &App{
		botMgr:      NewBotManager(),
		messages:    make([]ChatMessage, 0),
		sessions:    make([]SessionInfo, 0),
		channelMsgs: make([]ChannelMessage, 0),
		gwLogs:      make([]string, 0),
	}
}

// activeConfigPath returns the active bot's config file path
func (a *App) activeConfigPath() string {
	bot := a.botMgr.GetActiveBot()
	if bot != nil {
		return bot.ConfigPath
	}
	// Fallback to standard config dir
	return filepath.Join(configDir(), "bots", "default", "config.json")
}

// activeWorkspace returns the active bot's workspace path
func (a *App) activeWorkspace() string {
	bot := a.botMgr.GetActiveBot()
	if bot != nil {
		return bot.Workspace
	}
	return filepath.Join(configDir(), "bots", "default", "workspace")
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
	wailsRuntime.EventsEmit(a.ctx, "bots:updated", a.ListBots())
}

func (a *App) shutdown(ctx context.Context) {
	a.StopGateway()
	a.stopWSBridge()
}

func (a *App) beforeClose(ctx context.Context) (prevent bool) {
	a.HideWindow()
	return true
}

func (a *App) watchSignals() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)
	<-sigChan
	a.StopGateway()
	a.stopWSBridge()
	os.Exit(0)
}

func (a *App) autoStartGateway() {
	time.Sleep(1 * time.Second)
	configPath := a.activeConfigPath()
	data, err := os.ReadFile(configPath)
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
				err := a.StartGateway()
				if err == nil {
					a.addLog("[auto] Gateway auto-started (enabled channels detected)")
				}
				return
			}
		}
	}
}

// ==================== Bot Management (Wails bindings) ====================

func (a *App) ListBots() []map[string]interface{} {
	bots := a.botMgr.ListBots()
	activeID := a.botMgr.GetActiveBotID()
	result := make([]map[string]interface{}, 0, len(bots))
	for _, b := range bots {
		item := map[string]interface{}{
			"id":        b.ID,
			"name":      b.Name,
			"avatar":    b.Avatar,
			"port":      b.Port,
			"isActive":  b.ID == activeID,
			"createdAt": b.CreatedAt,
			"running":   false,
		}
		// Check if this bot's gateway is running
		a.gatewayMu.Lock()
		if a.gwStatus.Running && a.gwStatus.BotID == b.ID {
			item["running"] = true
			item["pid"] = a.gwStatus.PID
		}
		a.gatewayMu.Unlock()
		result = append(result, item)
	}
	return result
}

func (a *App) GetActiveBot() map[string]interface{} {
	bot := a.botMgr.GetActiveBot()
	if bot == nil {
		return nil
	}
	return map[string]interface{}{
		"id":       bot.ID,
		"name":     bot.Name,
		"avatar":   bot.Avatar,
		"port":     bot.Port,
		"isActive": true,
	}
}

func (a *App) CreateBot(name, avatar string) (map[string]interface{}, error) {
	bot, err := a.botMgr.CreateBot(name, avatar)
	if err != nil {
		return nil, err
	}
	wailsRuntime.EventsEmit(a.ctx, "bots:updated", a.ListBots())
	return map[string]interface{}{
		"id":     bot.ID,
		"name":   bot.Name,
		"avatar": bot.Avatar,
		"port":   bot.Port,
	}, nil
}

func (a *App) DeleteBot(id string) error {
	// Stop gateway first if this bot is active
	a.gatewayMu.Lock()
	if a.gwStatus.BotID == id && a.gwStatus.Running {
		a.gatewayMu.Unlock()
		a.StopGateway()
	} else {
		a.gatewayMu.Unlock()
	}
	err := a.botMgr.DeleteBot(id)
	if err != nil {
		return err
	}
	wailsRuntime.EventsEmit(a.ctx, "bots:updated", a.ListBots())
	return nil
}

func (a *App) UpdateBot(id, name, avatar string) error {
	_, err := a.botMgr.UpdateBot(id, name, avatar)
	if err != nil {
		return err
	}
	wailsRuntime.EventsEmit(a.ctx, "bots:updated", a.ListBots())
	return nil
}

func (a *App) SwitchBot(id string) error {
	// Stop current gateway
	a.StopGateway()

	err := a.botMgr.SetActiveBot(id)
	if err != nil {
		return err
	}

	// Clear logs, messages, sessions for new bot
	a.gwLogsMu.Lock()
	a.gwLogs = make([]string, 0)
	a.gwLogsMu.Unlock()

	a.msgMu.Lock()
	a.messages = make([]ChatMessage, 0)
	a.msgMu.Unlock()

	a.sessionMu.Lock()
	a.sessions = make([]SessionInfo, 0)
	a.sessionMu.Unlock()

	a.channelMsgMu.Lock()
	a.channelMsgs = make([]ChannelMessage, 0)
	a.channelMsgMu.Unlock()

	// Emit updates
	wailsRuntime.EventsEmit(a.ctx, "bots:updated", a.ListBots())
	wailsRuntime.EventsEmit(a.ctx, "bot:switched", id)
	wailsRuntime.EventsEmit(a.ctx, "gateway:status", a.GetGatewayStatus())

	// Auto-start new bot's gateway
	go a.autoStartGateway()

	return nil
}

func (a *App) GetBotConfig(id string) string {
	path := a.botMgr.GetBotConfigPath(id)
	if path == "" {
		return "{}"
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "{}"
	}
	return string(data)
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

func (a *App) ShowAndNavigate(page string) {
	a.ShowWindow()
	a.NavigateTo(page)
}

// ==================== Config Management ====================

func (a *App) GetConfig() string {
	configPath := a.activeConfigPath()
	data, err := os.ReadFile(configPath)
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

	configPath := a.activeConfigPath()

	backup := configPath + ".bak"
	if data, err := os.ReadFile(configPath); err == nil {
		os.WriteFile(backup, data, 0600)
	}
	formatted, err := formatJSON(configJSON)
	if err != nil {
		return fmt.Errorf("failed to format JSON: %w", err)
	}
	if err := os.WriteFile(configPath, []byte(formatted), 0600); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}
	wailsRuntime.EventsEmit(a.ctx, "config:saved", true)
	return nil
}

// ==================== Provider Management ====================

func (a *App) SetProviderField(provider string, field string, value string) error {
	configPath := a.activeConfigPath()
	data, err := os.ReadFile(configPath)
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
	var parsed interface{}
	if err := json.Unmarshal([]byte(value), &parsed); err != nil {
		parsed = value
	}
	p[field] = parsed
	providers[provider] = p
	newData, _ := json.MarshalIndent(raw, "", "  ")
	if err := os.WriteFile(configPath, newData, 0600); err != nil {
		return err
	}
	wailsRuntime.EventsEmit(a.ctx, "config:saved", true)
	return nil
}

func (a *App) SetProviderAPIKey(provider string, apiKey string) error {
	return a.SetProviderField(provider, "apiKey", apiKey)
}

func (a *App) GetProviders() map[string]interface{} {
	configPath := a.activeConfigPath()
	data, err := os.ReadFile(configPath)
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

// ==================== Agent Defaults ====================

func (a *App) GetAgentDefaults() map[string]interface{} {
	configPath := a.activeConfigPath()
	data, err := os.ReadFile(configPath)
	if err != nil {
		return map[string]interface{}{}
	}
	var raw map[string]interface{}
	json.Unmarshal(data, &raw)
	agents, ok := raw["agents"].(map[string]interface{})
	if !ok {
		return map[string]interface{}{}
	}
	defaults, ok := agents["defaults"].(map[string]interface{})
	if !ok {
		return map[string]interface{}{}
	}
	return defaults
}

func (a *App) SetAgentDefaults(defaultsJSON string) error {
	configPath := a.activeConfigPath()
	data, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	agents, ok := raw["agents"].(map[string]interface{})
	if !ok {
		agents = make(map[string]interface{})
		raw["agents"] = agents
	}
	var defaults map[string]interface{}
	if err := json.Unmarshal([]byte(defaultsJSON), &defaults); err != nil {
		return err
	}
	agents["defaults"] = defaults
	newData, _ := json.MarshalIndent(raw, "", "  ")
	if err := os.WriteFile(configPath, newData, 0600); err != nil {
		return err
	}
	wailsRuntime.EventsEmit(a.ctx, "config:saved", true)
	return nil
}

// ==================== Channel Management ====================

func (a *App) GetChannels() map[string]interface{} {
	configPath := a.activeConfigPath()
	data, err := os.ReadFile(configPath)
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
	configPath := a.activeConfigPath()
	data, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}
	var raw map[string]interface{}
	json.Unmarshal(data, &raw)
	channels, ok := raw["channels"].(map[string]interface{})
	if !ok {
		channels = make(map[string]interface{})
		raw["channels"] = channels
	}
	ch, ok := channels[channel].(map[string]interface{})
	if !ok {
		ch = make(map[string]interface{})
	}
	var parsed interface{}
	if err := json.Unmarshal([]byte(value), &parsed); err != nil {
		parsed = value
	}
	ch[field] = parsed
	newData, _ := json.MarshalIndent(raw, "", "  ")
	if err := os.WriteFile(configPath, newData, 0600); err != nil {
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

	bot := a.botMgr.GetActiveBot()
	if bot == nil {
		return fmt.Errorf("no active bot")
	}

	configPath := bot.ConfigPath
	port := bot.Port

	// Allow config override
	if data, err := os.ReadFile(configPath); err == nil {
		var raw map[string]interface{}
		json.Unmarshal(data, &raw)
		if gw, ok := raw["gateway"].(map[string]interface{}); ok {
			if p, ok := gw["port"].(float64); ok && int(p) != 0 {
				port = int(p)
			}
		}
	}

	args := []string{"gateway", "--port", fmt.Sprintf("%d", port), "--config", configPath}
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
		ConfigPath: configPath,
		BotID:      bot.ID,
	}

	a.emitGatewayStatus()
	wailsRuntime.EventsEmit(a.ctx, "gateway:started", true)
	a.addLog(fmt.Sprintf("[gateway] Bot %q started on port %d (PID %d)", bot.Name, port, cmd.Process.Pid))

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
		wailsRuntime.EventsEmit(a.ctx, "bots:updated", a.ListBots())
	}()

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
	wailsRuntime.EventsEmit(a.ctx, "bots:updated", a.ListBots())
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

func (a *App) httpPollLoop(port int) {
	baseURL := fmt.Sprintf("http://127.0.0.1:%d", port)
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		if !a.gwStatus.Running {
			return
		}
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

	configPath := a.activeConfigPath()

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
		args := []string{"agent", "--message", message, "--config", configPath}
		cmd := exec.Command(nanobotBin, args...)

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

		scanner := bufio.NewScanner(stdout)
		scanner.Buffer(make([]byte, 64*1024), 1024*1024)
		var fullContent strings.Builder
		for scanner.Scan() {
			line := scanner.Text()
			fullContent.WriteString(line)
			fullContent.WriteString("\n")

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
	sessionsDir := filepath.Join(a.activeWorkspace(), "sessions")

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

func (a *App) DeleteSession(sessionPath string) error {
	if sessionPath == "" {
		return fmt.Errorf("session path is empty")
	}
	// Validate path is under workspace/sessions
	ws := a.activeWorkspace()
	absPath, err := filepath.Abs(sessionPath)
	if err != nil {
		return fmt.Errorf("invalid path: %w", err)
	}
	if !strings.HasPrefix(absPath, filepath.Join(ws, "sessions")) {
		return fmt.Errorf("path is outside sessions directory")
	}
	if err := os.Remove(absPath); err != nil {
		return fmt.Errorf("failed to delete: %w", err)
	}
	a.scanSessions()
	return nil
}

// ==================== System Info ====================

func (a *App) GetSystemInfo() map[string]string {
	nanobotBin, _ := findNanobot()
	info := map[string]string{
		"os":              goRuntime.GOOS,
		"arch":            goRuntime.GOARCH,
		"nanobot":         nanobotBin,
		"configPath":      a.activeConfigPath(),
		"workspace":       a.activeWorkspace(),
		"configDir":       configDir(),
		"goVersion":       goRuntime.Version(),
		"version":         version,
		"channelMessages": fmt.Sprintf("%d", len(a.channelMsgs)),
		"sessions":        fmt.Sprintf("%d", len(a.sessions)),
		"totalBots":       fmt.Sprintf("%d", len(a.botMgr.ListBots())),
	}

	data, err := os.ReadFile(a.activeConfigPath())
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
	// 0. Check custom path from settings
	settingsPath := filepath.Join(configDir(), "settings.json")
	if data, err := os.ReadFile(settingsPath); err == nil {
		var settings map[string]interface{}
		json.Unmarshal(data, &settings)
		if cp, ok := settings["nanobotPath"].(string); ok && cp != "" {
			if _, err := os.Stat(cp); err == nil {
				return cp, nil
			}
		}
	}

	// 1. Check standard install directory (~/.local/share/hey-nanobot/bin/)
	home, _ := os.UserHomeDir()
	if home != "" {
		binName := "nanobot"
		if goRuntime.GOOS == "windows" {
			binName = "nanobot.exe"
		}
		standardPath := filepath.Join(home, ".local", "share", "hey-nanobot", "bin", binName)
		if _, err := os.Stat(standardPath); err == nil {
			return standardPath, nil
		}
	}

	// 2. Check bundled binary (legacy — next to the executable)
	if exe, err := os.Executable(); err == nil {
		bundled := filepath.Join(filepath.Dir(exe), "nanobot-bin", "nanobot")
		if goRuntime.GOOS == "windows" {
			bundled += ".exe"
		}
		if _, err := os.Stat(bundled); err == nil {
			return bundled, nil
		}
		// Also check Resources dir (macOS .app bundle)
		resources := filepath.Join(filepath.Dir(exe), "..", "Resources", "nanobot-bin", "nanobot")
		if _, err := os.Stat(resources); err == nil {
			return resources, nil
		}
	}

	// 3. Check PATH
	if path, err := exec.LookPath("nanobot"); err == nil {
		return path, nil
	}

	// 4. Check common install locations
	candidates := []string{
		filepath.Join(home, ".local", "bin", "nanobot"),
	}
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return c, nil
		}
	}
	return "", fmt.Errorf("nanobot not found")
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

type LineWriter struct {
	ctx    context.Context
	event  string
	onLine func(string)
	buf    []byte
	bufMu  sync.Mutex
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
	wailsRuntime.EventsEmit(w.ctx, w.event, string(p))
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

// ==================== Setup Wizard Bindings ====================

// CheckNanobotInstalled checks if nanobot binary is available
func (a *App) CheckNanobotInstalled() string {
	path, err := findNanobot()
	if err != nil {
		return ""
	}
	return path
}

// SetNanobotPath allows user to specify a custom nanobot binary path.
// Saves to app settings file (not config.json).
func (a *App) SetNanobotPath(customPath string) error {
	settingsPath := filepath.Join(configDir(), "settings.json")
	settings := map[string]interface{}{}
	if data, err := os.ReadFile(settingsPath); err == nil {
		json.Unmarshal(data, &settings)
	}
	settings["nanobotPath"] = customPath
	data, _ := json.MarshalIndent(settings, "", "  ")
	return os.WriteFile(settingsPath, data, 0644)
}

// GetNanobotInfo returns detailed info about the nanobot binary being used
func (a *App) GetNanobotInfo() map[string]interface{} {
	result := map[string]interface{}{
		"path":       "",
		"source":     "none",
		"version":    "",
		"available":  false,
		"customPath": "",
	}

	// Check custom path first
	settingsPath := filepath.Join(configDir(), "settings.json")
	if data, err := os.ReadFile(settingsPath); err == nil {
		var settings map[string]interface{}
		json.Unmarshal(data, &settings)
		if cp, ok := settings["nanobotPath"].(string); ok && cp != "" {
			result["customPath"] = cp
			if _, err := os.Stat(cp); err == nil {
				result["path"] = cp
				result["source"] = "custom"
				result["available"] = true
			}
		}
	}

	// If no custom path or custom path invalid, find normally
	if !result["available"].(bool) {
		if path, err := findNanobot(); err == nil {
			result["path"] = path
			result["available"] = true

			// Determine source
			home, _ := os.UserHomeDir()
			binName := "nanobot"
			if goRuntime.GOOS == "windows" {
				binName = "nanobot.exe"
			}
			standardPath := filepath.Join(home, ".local", "share", "hey-nanobot", "bin", binName)
			if path == standardPath {
				result["source"] = "standard"
			}

			// Check bundled (legacy)
			if result["source"] == "none" {
				if exe, err := os.Executable(); err == nil {
					bundled := filepath.Join(filepath.Dir(exe), "nanobot-bin", binName)
					if path == bundled {
						result["source"] = "bundled"
					} else {
						resources := filepath.Join(filepath.Dir(exe), "..", "Resources", "nanobot-bin", binName)
						if path == resources {
							result["source"] = "bundled"
						}
					}
				}
			}

			if result["source"] == "none" {
				result["source"] = "external"
			}
		}
	}

	// Try to get version
	if result["available"].(bool) {
		path := result["path"].(string)
		if out, err := exec.Command(path, "--version").CombinedOutput(); err == nil {
			result["version"] = strings.TrimSpace(string(out))
		}
	}

	return result
}

// GetSetupState returns whether setup is needed
// Note: nanobot binary is NOT required for setup — bundled binary is preferred.
func (a *App) GetSetupState() map[string]interface{} {
	result := map[string]interface{}{
		"needsSetup":  false,
		"nanobotPath": "",
		"hasProvider": false,
		"hasChannel":  false,
	}

	// Check nanobot (informational only — not a setup blocker)
	if path, err := findNanobot(); err == nil {
		result["nanobotPath"] = path
	}

	// Check config
	configPath := a.activeConfigPath()
	data, err := os.ReadFile(configPath)
	if err != nil {
		// No config file at all → needs setup
		result["needsSetup"] = true
		return result
	}
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		result["needsSetup"] = true
		return result
	}

	// Check providers — need at least one API key
	if providers, ok := raw["providers"].(map[string]interface{}); ok {
		for _, p := range providers {
			if cfg, ok := p.(map[string]interface{}); ok {
				if key, ok := cfg["api_key"].(string); ok && key != "" {
					result["hasProvider"] = true
					break
				}
				if key, ok := cfg["apiKey"].(string); ok && key != "" {
					result["hasProvider"] = true
					break
				}
			}
		}
	}
	if !result["hasProvider"].(bool) {
		result["needsSetup"] = true
	}

	// Channel is optional — don't block setup on it
	if channels, ok := raw["channels"].(map[string]interface{}); ok {
		for _, ch := range channels {
			if cfg, ok := ch.(map[string]interface{}); ok {
				if enabled, ok := cfg["enabled"].(bool); ok && enabled {
					result["hasChannel"] = true
					break
				}
			}
		}
	}

	return result
}

// SetupSaveConfig saves config during setup wizard (creates if needed)
func (a *App) SetupSaveConfig(providersJson string, channelsJson string) error {
	configPath := a.activeConfigPath()

	// Ensure directory exists
	os.MkdirAll(filepath.Dir(configPath), 0755)

	// Load existing or create new
	raw := map[string]interface{}{}
	if data, err := os.ReadFile(configPath); err == nil {
		json.Unmarshal(data, &raw)
	}

	// Merge providers
	if providersJson != "" {
		var providers map[string]interface{}
		if err := json.Unmarshal([]byte(providersJson), &providers); err == nil {
			if raw["providers"] == nil {
				raw["providers"] = map[string]interface{}{}
			}
			if existing, ok := raw["providers"].(map[string]interface{}); ok {
				for k, v := range providers {
					existing[k] = v
				}
			}
		}
	}

	// Merge channels
	if channelsJson != "" {
		var channels map[string]interface{}
		if err := json.Unmarshal([]byte(channelsJson), &channels); err == nil {
			if raw["channels"] == nil {
				raw["channels"] = map[string]interface{}{}
			}
			if existing, ok := raw["channels"].(map[string]interface{}); ok {
				for k, v := range channels {
					existing[k] = v
				}
			}
		}
	}

	// Ensure model field
	if raw["model"] == nil {
		raw["model"] = "gpt-4o-mini"
	}

	data, _ := json.MarshalIndent(raw, "", "  ")
	return os.WriteFile(configPath, data, 0644)
}
