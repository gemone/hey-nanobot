#!/usr/bin/env bash
# build.sh — Cross-platform build script for Hey Nanobot
# Usage:
#   ./build.sh              # Build current platform
#   ./build.sh darwin       # Build macOS arm64
#   ./build.sh windows      # Build Windows amd64
#   ./build.sh linux        # Build Linux amd64
#   ./build.sh all          # Build all 3 platforms
#   ./build.sh clean        # Clean build artifacts

set -euo pipefail
cd "$(dirname "$0")"

APP_NAME="hey-nanobot"
VERSION="0.1.0"
BUILD_DIR="build/bin"
LDFLAGS="-s -w -X main.version=${VERSION}"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m'

log()  { echo -e "${GREEN}[build]${NC} $*"; }
warn() { echo -e "${YELLOW}[warn]${NC} $*"; }
err()  { echo -e "${RED}[error]${NC} $*" >&2; }

check_wails() {
    if ! command -v wails &>/dev/null && [ ! -f "$HOME/go/bin/wails" ]; then
        err "Wails CLI not found. Install: go install github.com/wailsapp/wails/v2/cmd/wails@latest"
        exit 1
    fi
    WAILS_CMD="${HOME}/go/bin/wails"
    [ -f "$WAILS_CMD" ] || WAILS_CMD="wails"
    log "Using Wails: $($WAILS_CMD version 2>/dev/null | head -1)"
}

build_platform() {
    local os="$1"
    local arch="${2:-amd64}"

    log "Building ${CYAN}${APP_NAME} v${VERSION}${NC} for ${YELLOW}${os}/${arch}${NC}..."

    case "$os" in
        darwin)
            GOOS=darwin GOARCH=$arch $WAILS_CMD build -ldflags "$LDFLAGS" -platform "darwin/$arch"
            local out="${BUILD_DIR}/${APP_NAME}.app"
            if [ -d "$out" ]; then
                local size=$(du -sh "$out" | cut -f1)
                log "✅ macOS ($arch): $out ($size)"
            fi
            ;;
        windows)
            GOOS=windows GOARCH=$arch $WAILS_CMD build -ldflags "$LDFLAGS" -platform "windows/$arch"
            local out="${BUILD_DIR}/${APP_NAME}.exe"
            if [ -f "$out" ]; then
                local size=$(ls -lh "$out" | awk '{print $5}')
                log "✅ Windows ($arch): $out ($size)"
            fi
            ;;
        linux)
            GOOS=linux GOARCH=$arch $WAILS_CMD build -ldflags "$LDFLAGS" -platform "linux/$arch"
            local out="${BUILD_DIR}/${APP_NAME}"
            if [ -f "$out" ]; then
                local size=$(ls -lh "$out" | awk '{print $5}')
                # Copy desktop entry
                cp -f build/linux/hey-nanobot.desktop "${BUILD_DIR}/" 2>/dev/null || true
                log "✅ Linux ($arch): $out ($size)"
            fi
            ;;
        *)
            err "Unknown platform: $os"
            return 1
            ;;
    esac
}

build_all() {
    log "Building all platforms..."
    echo ""

    # macOS arm64 (Apple Silicon)
    build_platform darwin arm64 "darwin-arm64"

    # macOS amd64 (Intel)
    build_platform darwin amd64 "darwin-amd64"

    # Windows amd64
    build_platform windows amd64 "windows-amd64"

    # Linux amd64
    build_platform linux amd64 "linux-amd64"

    echo ""
    log "All builds complete!"
    log "Output directory: ${BUILD_DIR}/"
    ls -lh "${BUILD_DIR}/" 2>/dev/null | tail -n +2
}

clean() {
    log "Cleaning build artifacts..."
    rm -rf "${BUILD_DIR:?}"/*
    log "✅ Cleaned"
}

# ====== Main ======

check_wails

case "${1:-current}" in
    darwin|macos)
        build_platform darwin arm64
        ;;
    darwin-intel|macos-intel)
        build_platform darwin amd64
        ;;
    windows|win)
        build_platform windows amd64
        ;;
    linux)
        build_platform linux amd64
        ;;
    all)
        build_all
        ;;
    clean)
        clean
        ;;
    current|"")
        # Build for current OS
        case "$(uname -s)" in
            Darwin) build_platform darwin arm64 ;;
            Linux)  build_platform linux amd64 ;;
            MINGW*|MSYS*|CYGWIN*) build_platform windows amd64 ;;
            *) err "Unknown OS: $(uname -s)"; exit 1 ;;
        esac
        ;;
    *)
        echo "Usage: $0 {darwin|darwin-intel|windows|linux|all|clean}"
        echo ""
        echo "Platforms:"
        echo "  darwin          macOS Apple Silicon (arm64)"
        echo "  darwin-intel    macOS Intel (amd64)"
        echo "  windows         Windows (amd64)"
        echo "  linux           Linux (amd64)"
        echo "  all             Build all platforms"
        echo "  clean           Remove build artifacts"
        exit 1
        ;;
esac
