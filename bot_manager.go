package main

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// ==================== Bot Data Types ====================

// Bot represents a single bot instance configuration
type Bot struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	ConfigPath string `json:"configPath"`
	Workspace  string `json:"workspace"`
	Port       int    `json:"port"`
	Avatar     string `json:"avatar"`
	CreatedAt  string `json:"createdAt"`
}

// BotRegistry is the persisted registry of all bots
type BotRegistry struct {
	Version     int             `json:"version"`
	ActiveBotID string          `json:"activeBotId"`
	Bots        map[string]*Bot `json:"bots"`
	NextPort    int             `json:"nextPort"`
}

// BotStatus combines bot info with its runtime status
type BotStatus struct {
	Bot
	Running   bool   `json:"running"`
	PID       int    `json:"pid"`
	Port      int    `json:"port"`
	Uptime    string `json:"uptime"`
	IsActive  bool   `json:"isActive"`
}

// ==================== BotManager ====================

const (
	appDirName = "hey-nanobot"
	basePort   = 18790
)

var defaultAvatars = []string{"🐱", "🤖", "🦊", "🐶", "🐸", "🦁", "🐼", "🦄", "🐲", "👻", "🧠", "⚡"}

type BotManager struct {
	mu           sync.RWMutex
	registryPath string
	appDir       string // Standard config dir (e.g. ~/Library/Application Support/hey-nanobot)
	registry     *BotRegistry
}

// configDir returns the OS-standard config directory.
// macOS: ~/Library/Application Support/hey-nanobot
// Linux: ~/.config/hey-nanobot
// Windows: %AppData%/hey-nanobot
func configDir() string {
	if dir, err := os.UserConfigDir(); err == nil {
		return filepath.Join(dir, appDirName)
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, "."+appDirName)
}

// ==================== Constructor ====================

func NewBotManager() *BotManager {
	appDir := configDir()

	bm := &BotManager{
		appDir:       appDir,
		registryPath: filepath.Join(appDir, "registry.json"),
	}

	bm.loadOrInit()
	return bm
}

func (bm *BotManager) loadOrInit() {
	// Try to load existing registry
	data, err := os.ReadFile(bm.registryPath)
	if err == nil {
		var reg BotRegistry
		if json.Unmarshal(data, &reg) == nil && reg.Version == 1 && len(reg.Bots) > 0 {
			bm.registry = &reg
			return
		}
	}

	// Initialize new registry
	bm.registry = &BotRegistry{
		Version:  1,
		Bots:     make(map[string]*Bot),
		NextPort: basePort + 1,
	}

	// Create a fresh default bot
	bm.createBotLocked("Default Bot", "🐱")
	bm.saveLocked()
}

// ==================== CRUD ====================

func (bm *BotManager) CreateBot(name, avatar string) (*Bot, error) {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	if avatar == "" {
		avatar = defaultAvatars[len(bm.registry.Bots)%len(defaultAvatars)]
	}

	bot := bm.createBotLocked(name, avatar)
	if err := bm.saveLocked(); err != nil {
		return nil, err
	}
	return bot, nil
}

func (bm *BotManager) createBotLocked(name, avatar string) *Bot {
	id := generateBotID()
	botDir := filepath.Join(bm.appDir, "bots", id)
	port := bm.registry.NextPort

	bot := &Bot{
		ID:         id,
		Name:       name,
		ConfigPath: filepath.Join(botDir, "config.json"),
		Workspace:  filepath.Join(botDir, "workspace"),
		Port:       port,
		Avatar:     avatar,
		CreatedAt:  time.Now().Format("2006-01-02T15:04:05"),
	}

	bm.registry.Bots[id] = bot
	bm.registry.NextPort = port + 1

	if bm.registry.ActiveBotID == "" {
		bm.registry.ActiveBotID = id
	}

	// Create directories
	os.MkdirAll(bot.Workspace, 0755)
	os.MkdirAll(filepath.Dir(bot.ConfigPath), 0755)

	// Init default config if not exists
	bm.initBotConfig(bot)

	return bot
}

func (bm *BotManager) DeleteBot(id string) error {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	if len(bm.registry.Bots) <= 1 {
		return fmt.Errorf("cannot delete the last bot")
	}

	bot, ok := bm.registry.Bots[id]
	if !ok {
		return fmt.Errorf("bot %s not found", id)
	}

	// Remove from registry
	delete(bm.registry.Bots, id)

	// If this was active, switch to another
	if bm.registry.ActiveBotID == id {
		for k := range bm.registry.Bots {
			bm.registry.ActiveBotID = k
			break
		}
	}

	// Remove config dir
	if bot.ConfigPath != "" {
		os.RemoveAll(filepath.Dir(bot.ConfigPath))
	}

	return bm.saveLocked()
}

func (bm *BotManager) UpdateBot(id, name, avatar string) (*Bot, error) {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	bot, ok := bm.registry.Bots[id]
	if !ok {
		return nil, fmt.Errorf("bot %s not found", id)
	}
	if name != "" {
		bot.Name = name
	}
	if avatar != "" {
		bot.Avatar = avatar
	}

	if err := bm.saveLocked(); err != nil {
		return nil, err
	}
	return bot, nil
}

// ==================== Active Bot ====================

func (bm *BotManager) GetActiveBotID() string {
	bm.mu.RLock()
	defer bm.mu.RUnlock()
	return bm.registry.ActiveBotID
}

func (bm *BotManager) GetActiveBot() *Bot {
	bm.mu.RLock()
	defer bm.mu.RUnlock()
	return bm.registry.Bots[bm.registry.ActiveBotID]
}

func (bm *BotManager) SetActiveBot(id string) error {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	if _, ok := bm.registry.Bots[id]; !ok {
		return fmt.Errorf("bot %s not found", id)
	}
	bm.registry.ActiveBotID = id
	return bm.saveLocked()
}

// ==================== Queries ====================

func (bm *BotManager) ListBots() []*Bot {
	bm.mu.RLock()
	defer bm.mu.RUnlock()

	bots := make([]*Bot, 0, len(bm.registry.Bots))
	for _, b := range bm.registry.Bots {
		bots = append(bots, b)
	}
	return bots
}

func (bm *BotManager) GetBot(id string) (*Bot, error) {
	bm.mu.RLock()
	defer bm.mu.RUnlock()

	bot, ok := bm.registry.Bots[id]
	if !ok {
		return nil, fmt.Errorf("bot %s not found", id)
	}
	return bot, nil
}

func (bm *BotManager) GetBotConfigPath(id string) string {
	bm.mu.RLock()
	defer bm.mu.RUnlock()
	if bot, ok := bm.registry.Bots[id]; ok {
		return bot.ConfigPath
	}
	return ""
}

// ==================== Persistence ====================

func (bm *BotManager) saveLocked() error {
	data, err := json.MarshalIndent(bm.registry, "", "  ")
	if err != nil {
		return err
	}
	os.MkdirAll(filepath.Dir(bm.registryPath), 0755)
	return os.WriteFile(bm.registryPath, data, 0644)
}

func (bm *BotManager) initBotConfig(bot *Bot) {
	if _, err := os.Stat(bot.ConfigPath); err == nil {
		return
	}

	defaultCfg := map[string]interface{}{
		"agents": map[string]interface{}{
			"defaults": map[string]interface{}{
				"workspace":              bot.Workspace,
				"model":                  "anthropic/claude-sonnet-4-20250514",
				"provider":               "auto",
				"maxTokens":              8192,
				"contextWindowTokens":    65536,
				"temperature":            0.1,
				"maxToolIterations":      40,
			},
		},
		"gateway": map[string]interface{}{
			"port": bot.Port,
		},
		"channels": map[string]interface{}{
			"sendProgress": true,
		},
		"providers": map[string]interface{}{},
		"tools":     map[string]interface{}{},
	}

	data, _ := json.MarshalIndent(defaultCfg, "", "  ")
	os.WriteFile(bot.ConfigPath, data, 0644)
}

// ==================== Helpers ====================

func generateBotID() string {
	// 8-char hex ID
	buf := make([]byte, 4)
	_, _ = rand.Read(buf)
	return fmt.Sprintf("%08x", buf)
}
