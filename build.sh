#!/usr/bin/env bash
# build.sh — Cross-platform build script for Hey Nanobot
# Usage:
#   ./build.sh              # Build current platform
#   ./build.sh darwin       # Build macOS arm64
#   ./build.sh windows      # Build Windows amd64
#   ./build.sh linux        # Build Linux amd64
#   ./build.sh all          # Build all 3 platforms
#   ./build.sh nanobot      # Build nanobot binary with PyInstaller
#   ./build.sh full         # Build nanobot + desktop app
#   ./build.sh clean        # Clean build artifacts

set -euo pipefail
cd "$(dirname "$0")"

APP_NAME="hey-nanobot"
VERSION="1.2.0"
BUILD_DIR="build/bin"
LDFLAGS="-s -w -X main.version=${VERSION}"
NANOBOT_SRC="nanobot-bin/nanobot"

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

# Embed nanobot binary into the build output
embed_nanobot() {
    local platform="$1"
    local target_dir="$2"

    local nano_bin="$NANOBOT_SRC"
    if [ "$platform" = "windows" ]; then
        nano_bin="${NANOBOT_SRC}.exe"
    fi

    if [ ! -f "$nano_bin" ]; then
        warn "nanobot binary not found at $nano_bin — skipping embed"
        warn "Run './build.sh nanobot' first to build it"
        return
    fi

    local dest="$target_dir/nanobot-bin"
    mkdir -p "$dest"
    cp -f "$nano_bin" "$dest/"
    chmod +x "$dest/"*
    log "Embedded nanobot binary → $dest/"
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
                embed_nanobot "darwin" "$out/Contents/Resources"
                local size=$(du -sh "$out" | cut -f1)
                log "✅ macOS ($arch): $out ($size)"
            fi
            ;;
        windows)
            GOOS=windows GOARCH=$arch $WAILS_CMD build -ldflags "$LDFLAGS" -platform "windows/$arch"
            local out="${BUILD_DIR}/${APP_NAME}.exe"
            if [ -f "$out" ]; then
                embed_nanobot "windows" "$(dirname "$out")"
                local size=$(ls -lh "$out" | awk '{print $5}')
                log "✅ Windows ($arch): $out ($size)"
            fi
            ;;
        linux)
            GOOS=linux GOARCH=$arch $WAILS_CMD build -ldflags "$LDFLAGS" -platform "linux/$arch"
            local out="${BUILD_DIR}/${APP_NAME}"
            if [ -f "$out" ]; then
                embed_nanobot "linux" "$BUILD_DIR"
                cp -f build/linux/hey-nanobot.desktop "${BUILD_DIR}/" 2>/dev/null || true
                local size=$(ls -lh "$out" | awk '{print $5}')
                log "✅ Linux ($arch): $out ($size)"
            fi
            ;;
        *)
            err "Unknown platform: $os"
            return 1
            ;;
    esac
}

build_nanobot() {
    log "Building nanobot binary with PyInstaller..."
    if [ ! -f "build_nano.py" ]; then
        err "build_nano.py not found"
        exit 1
    fi

    # Find nanobot-ai Python
    local nano_python="$HOME/.local/share/uv/tools/nanobot-ai/bin/python3"
    if [ ! -f "$nano_python" ]; then
        err "nanobot-ai Python not found at $nano_python"
        err "Install: uv tool install nanobot-ai"
        exit 1
    fi

    # Check pyinstaller
    if ! $nano_python -c "import PyInstaller" 2>/dev/null; then
        log "Installing PyInstaller into nanobot environment..."
        uv pip install --python "$nano_python" pyinstaller 2>&1 | tail -3
    fi

    $nano_python build_nano.py 2>&1

    if [ -f "$NANOBOT_SRC" ]; then
        local size=$(ls -lh "$NANOBOT_SRC" | awk '{print $5}')
        log "✅ nanobot binary: $NANOBOT_SRC ($size)"
    else
        err "PyInstaller build failed"
        exit 1
    fi
}

build_all() {
    log "Building all platforms..."
    echo ""

    # macOS arm64 (Apple Silicon)
    build_platform darwin arm64

    # macOS amd64 (Intel)
    build_platform darwin amd64

    # Windows amd64
    build_platform windows amd64

    # Linux amd64
    build_platform linux amd64

    echo ""
    log "All builds complete!"
    log "Output directory: ${BUILD_DIR}/"
    ls -lh "${BUILD_DIR}/" 2>/dev/null | tail -n +2
}

clean() {
    log "Cleaning build artifacts..."
    find build/bin -type f -delete 2>/dev/null
    find build/bin -type l -delete 2>/dev/null
    find build/bin -type d -empty -delete 2>/dev/null
    find build_nanobot -type f -delete 2>/dev/null
    find build_nanobot -type d -empty -delete 2>/dev/null
    log "✅ Cleaned"
}

# ====== Main ======

case "${1:-current}" in
    nanobot|nano)
        build_nanobot
        ;;
    darwin|macos)
        check_wails
        build_platform darwin arm64
        ;;
    darwin-intel|macos-intel)
        check_wails
        build_platform darwin amd64
        ;;
    windows|win)
        check_wails
        build_platform windows amd64
        ;;
    linux)
        check_wails
        build_platform linux amd64
        ;;
    all)
        check_wails
        build_all
        ;;
    full)
        build_nanobot
        check_wails
        build_platform darwin arm64
        ;;
    clean)
        clean
        ;;
    current|"")
        check_wails
        case "$(uname -s)" in
            Darwin) build_platform darwin arm64 ;;
            Linux)  build_platform linux amd64 ;;
            MINGW*|MSYS*|CYGWIN*) build_platform windows amd64 ;;
            *) err "Unknown OS: $(uname -s)"; exit 1 ;;
        esac
        ;;
    *)
        echo "Usage: $0 {nanobot|darwin|darwin-intel|windows|linux|all|full|clean}"
        echo ""
        echo "Commands:"
        echo "  nanobot          Build nanobot binary with PyInstaller"
        echo "  darwin           macOS Apple Silicon (arm64)"
        echo "  darwin-intel     macOS Intel (amd64)"
        echo "  windows          Windows (amd64)"
        echo "  linux            Linux (amd64)"
        echo "  all              Build all platforms"
        echo "  full             Build nanobot + macOS arm64 app"
        echo "  clean            Remove build artifacts"
        exit 1
        ;;
esac
