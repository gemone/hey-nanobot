<div align="center">

# 🐈 Hey Nanobot

**Personal AI Assistant Desktop App**

[中文](#中文) · [English](#english)

---

Built with Go · Wails · Vue 3 · TypeScript

[![Build](https://github.com/gemone/hey-nanobot/actions/workflows/build.yml/badge.svg)](https://github.com/gemone/hey-nanobot/actions/workflows/build.yml)
[![Go](https://img.shields.io/badge/Go-1.26-00ADD8?logo=go)](https://go.dev)
[![Wails](https://img.shields.io/badge/Wails-v2.12-6366f1)](https://wails.io)
[![Vue](https://img.shields.io/badge/Vue-3-4FC08D?logo=vue.js)](https://vuejs.org)

</div>

---

<a id="english"></a>

## 🇬🇧 English

### What is Hey Nanobot?

Hey Nanobot is a **desktop GUI** for [nanobot](https://github.com/HKUDS/nanobot) — an open-source AI assistant. It wraps the nanobot runtime with a native macOS/Windows/Linux experience, letting you manage bots, chat with AI, and monitor channels — all from one app.

### ✨ Features

| Category | Details |
|----------|---------|
| 🤖 **Multi-Bot Management** | Create, delete, switch between multiple independent bots |
| 💬 **Streaming Chat** | Real-time streaming AI responses with Markdown rendering |
| 📡 **Live Feed** | Real-time channel message stream (Telegram/Discord/QQ/etc.) |
| 🔗 **Channel Management** | Configure 20+ providers and 9+ messaging channels |
| 🔑 **Provider Management** | One-click API key configuration for OpenAI, Claude, Gemini, etc. |
| 🌐 **Gateway Control** | Start/stop/restart gateway with PID/port monitoring |
| 📂 **Session Management** | Browse and locate all chat session files |
| ⚙️ **Config Editor** | Raw JSON editor with format validation and backup |

### 🖼️ Tech Stack

```
┌─────────────────────────────────────────┐
│  Vue 3 + TypeScript + Vite (Frontend)   │
│              ↕ Wails Bindings            │
│  Go + Wails v2 (Backend + Native Shell) │
│              ↕ Process Management        │
│  nanobot (Python AI Runtime)            │
└─────────────────────────────────────────┘
```

- **Go** ~1,580 lines — Core logic, process management, WebSocket bridge
- **Vue/CSS** ~1,700 lines — Dark theme UI with purple accent
- **Python** — nanobot runtime (bundled via PyInstaller, 51MB standalone)

### 📦 Config Directory

Hey Nanobot stores config in the **OS-standard directory** and **never touches** `~/.nanobot/`:

| OS | Path |
|----|------|
| macOS | `~/Library/Application Support/hey-nanobot/` |
| Linux | `~/.config/hey-nanobot/` |
| Windows | `%AppData%/hey-nanobot/` |

Each bot has independent config and workspace:
```
hey-nanobot/
├── registry.json            # Bot registry (active bot, port allocation)
└── bots/
    └── {id}/                # Each bot:
        ├── config.json      #   Independent config (providers, channels)
        └── workspace/       #   Independent workspace
```

### 🚀 Quick Start

#### Prerequisites

- [Go](https://go.dev) 1.23+
- [Wails CLI](https://wails.io/docs/gettingstarted/installation/) v2.12+
- [Node.js](https://nodejs.org) 18+ (fnm/nvm recommended)
- [nanobot](https://github.com/HKUDS/nanobot) installed (`uv tool install nanobot-ai`)

#### Development

```bash
# Clone
git clone git@github.com:gemone/hey-nanobot.git
cd hey-nanobot

# Install frontend deps
cd frontend && npm install && cd ..

# Dev mode (hot reload)
wails dev
```

#### Build

```bash
# Quick build (macOS only, no nanobot embedding)
~/go/bin/wails build

# Full build with nanobot binary embedded
./build.sh full          # nanobot + macOS app
./build.sh darwin        # macOS arm64 (auto-embeds if available)
./build.sh darwin-intel  # macOS amd64
./build.sh windows       # Windows amd64
./build.sh all           # All platforms
```

Output: `build/bin/hey-nanobot.app` (macOS) or `build/bin/hey-nanobot.exe` (Windows)

### 🏗️ Project Structure

```
hey-nanobot/
├── main.go              # Entry point, window config, version
├── app.go               # Core logic (1,140 LOC)
├── bot_manager.go       # Multi-bot CRUD + registry (334 LOC)
├── build_nano.py        # PyInstaller build script for nanobot binary
├── build_nano.spec      # PyInstaller spec
├── build.sh             # Cross-platform build script
├── .github/workflows/
│   └── build.yml        # CI/CD (macOS/Windows/Linux)
└── frontend/src/
    ├── App.vue          # Main layout + bot switcher
    ├── style.css        # Dark theme (purple accent)
    └── components/
        ├── BotsPage.vue      # 🤖 Bot management
        ├── ChatPage.vue      # 💬 Streaming chat
        ├── FeedPage.vue      # 📡 Live channel feed
        ├── ChannelsPage.vue  # 🔗 Channel config
        ├── SessionsPage.vue  # 📂 Session browser
        ├── ProvidersPage.vue # 🔑 API keys
        ├── ConfigPage.vue    # ⚙️ JSON editor
        ├── GatewayPage.vue   # 🌐 Gateway control
        └── SystemPage.vue    # ℹ️ System info
```

### 🤝 Contributing

1. Fork → Branch → PR
2. `wails dev` for development
3. Test on your platform before submitting

### 📄 License

MIT

---

<a id="中文"></a>

## 🇨🇳 中文

### Hey Nanobot 是什么？

Hey Nanobot 是 [nanobot](https://github.com/HKUDS/nanobot)（开源 AI 助手）的**桌面客户端**。它将 nanobot 运行时封装为原生桌面应用，让你在一个窗口内管理多个 Bot、与 AI 对话、监控消息频道。

### ✨ 功能特性

| 分类 | 说明 |
|------|------|
| 🤖 **多 Bot 管理** | 创建、删除、切换多个独立 Bot，各 Bot 配置和 workspace 完全隔离 |
| 💬 **流式对话** | 实时流式 AI 回复，自动渲染 Markdown（代码块、列表、引用） |
| 📡 **实时消息流** | 实时查看 Telegram/Discord/QQ 等频道的消息流，颜色标签区分 |
| 🔗 **频道管理** | 配置 9+ 消息通道（Telegram、Discord、QQ、飞书、钉钉、微信企业等） |
| 🔑 **Provider 管理** | 一键配置 OpenAI、Claude、Gemini 等 20+ AI 模型的 API Key |
| 🌐 **网关控制** | 启动/停止/重启网关，实时显示 PID/端口/运行时间 |
| 📂 **会话管理** | 浏览所有聊天会话文件，一键定位 |
| ⚙️ **配置编辑** | 原始 JSON 编辑器，支持格式化和备份 |

### 🖥️ 技术栈

```
┌─────────────────────────────────────────┐
│  Vue 3 + TypeScript + Vite（前端）       │
│              ↕ Wails 绑定                │
│  Go + Wails v2（后端 + 原生窗口）        │
│              ↕ 进程管理                   │
│  nanobot（Python AI 运行时）             │
└─────────────────────────────────────────┘
```

- **Go** ~1,580 行 — 核心逻辑、进程管理、WebSocket 桥接
- **Vue/CSS** ~1,700 行 — 暗色主题 + 紫色强调色
- **Python** — nanobot 运行时（PyInstaller 打包为 51MB 独立二进制）

### 📦 配置目录

Hey Nanobot 的配置存放在 **操作系统标准目录**，**绝不触碰** `~/.nanobot/`：

| 系统 | 路径 |
|------|------|
| macOS | `~/Library/Application Support/hey-nanobot/` |
| Linux | `~/.config/hey-nanobot/` |
| Windows | `%AppData%/hey-nanobot/` |

每个 Bot 拥有独立的配置和工作空间：
```
hey-nanobot/
├── registry.json            # Bot 注册表（活跃 Bot、端口分配）
└── bots/
    └── {id}/                # 每个 Bot：
        ├── config.json      #   独立配置（模型、频道）
        └── workspace/       #   独立工作空间
```

### 🚀 快速开始

#### 环境要求

- [Go](https://go.dev) 1.23+
- [Wails CLI](https://wails.io/docs/gettingstarted/installation/) v2.12+
- [Node.js](https://nodejs.org) 18+（推荐 fnm/nvm）
- [nanobot](https://github.com/HKUDS/nanobot) 已安装（`uv tool install nanobot-ai`）

#### 开发模式

```bash
# 克隆
git clone git@github.com:gemone/hey-nanobot.git
cd hey-nanobot

# 安装前端依赖
cd frontend && npm install && cd ..

# 开发模式（热重载）
wails dev
```

#### 构建发布

```bash
# 快速构建（仅 macOS，不嵌入 nanobot）
~/go/bin/wails build

# 完整构建（嵌入 nanobot 二进制）
./build.sh full          # nanobot + macOS 应用
./build.sh darwin        # macOS arm64（自动嵌入）
./build.sh darwin-intel  # macOS amd64
./build.sh windows       # Windows amd64
./build.sh all           # 全平台
```

输出：`build/bin/hey-nanobot.app`（macOS）或 `build/bin/hey-nanobot.exe`（Windows）

### 🏗️ 项目结构

```
hey-nanobot/
├── main.go              # 入口、窗口配置、版本号
├── app.go               # 核心逻辑（1,140 行）
├── bot_manager.go       # 多 Bot CRUD + 注册表（334 行）
├── build_nano.py        # nanobot 的 PyInstaller 构建脚本
├── build_nano.spec      # PyInstaller 规格文件
├── build.sh             # 跨平台构建脚本
├── .github/workflows/
│   └── build.yml        # CI/CD（macOS/Windows/Linux）
└── frontend/src/
    ├── App.vue          # 主布局 + Bot 切换器
    ├── style.css        # 暗色主题（紫色强调）
    └── components/
        ├── BotsPage.vue      # 🤖 Bot 管理
        ├── ChatPage.vue      # 💬 流式对话
        ├── FeedPage.vue      # 📡 实时消息流
        ├── ChannelsPage.vue  # 🔗 频道配置
        ├── SessionsPage.vue  # 📂 会话浏览
        ├── ProvidersPage.vue # 🔑 API 密钥
        ├── ConfigPage.vue    # ⚙️ JSON 编辑器
        ├── GatewayPage.vue   # 🌐 网关控制
        └── SystemPage.vue    # ℹ️ 系统信息
```

### 🤝 参与贡献

1. Fork → 新分支 → 提 PR
2. 开发时使用 `wails dev`
3. 提交前请在你的平台上测试

### 📄 许可证

MIT
