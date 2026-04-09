#!/usr/bin/env bash
# install.sh — Install nanobot binary from GitHub Release
#
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/gemone/hey-nanobot/main/install.sh | bash
#   ./install.sh                  # Install latest
#   ./install.sh v1.2.3           # Install specific version
#
# Installs nanobot binary to ~/.local/share/hey-nanobot/bin/
# This is the standard location that hey-nanobot desktop app checks first.

set -euo pipefail

REPO="gemone/hey-nanobot"
INSTALL_DIR="${HEY_NANOBOT_BIN:-$HOME/.local/share/hey-nanobot/bin}"
VERSION="${1:-latest}"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m'

log()  { echo -e "${GREEN}[install]${NC} $*"; }
warn() { echo -e "${YELLOW}[warn]${NC} $*"; }
err()  { echo -e "${RED}[error]${NC} $*" >&2; }

# Detect platform
detect_platform() {
    local os arch
    os="$(uname -s | tr '[:upper:]' '[:lower:]')"
    arch="$(uname -m)"

    case "$os" in
        darwin) os="macos" ;;
        linux)  os="linux" ;;
        mingw*|msys*|cygwin*) os="windows" ;;
        *) err "Unsupported OS: $os"; exit 1 ;;
    esac

    case "$arch" in
        x86_64|amd64)  arch="amd64" ;;
        arm64|aarch64) arch="arm64" ;;
        *) err "Unsupported arch: $arch"; exit 1 ;;
    esac

    echo "${os}-${arch}"
}

# Get latest release tag from GitHub API
get_latest_version() {
    local url="https://api.github.com/repos/${REPO}/releases/latest"
    local tag
    tag="$(curl -fsSL "$url" 2>/dev/null | grep '"tag_name"' | head -1 | sed -E 's/.*"tag_name":\s*"([^"]+)".*/\1/')"
    if [ -z "$tag" ]; then
        err "Failed to get latest version from GitHub"
        exit 1
    fi
    echo "$tag"
}

# Find the nanobot asset URL for current platform
find_asset_url() {
    local platform="$1"
    local version="$2"
    local url="https://api.github.com/repos/${REPO}/releases/tags/${version}"

    local json
    json="$(curl -fsSL "$url" 2>/dev/null)"

    if [ -z "$json" ]; then
        err "Failed to fetch release info for ${version}"
        exit 1
    fi

    # Look for nanobot-{platform} asset
    local asset_url asset_name
    # macOS arm64
    for pattern in "nanobot-${platform}.tar.gz" "nanobot-${platform}.zip"; do
        asset_url="$(echo "$json" | python3 -c "
import sys, json
data = json.load(sys.stdin)
for asset in data.get('assets', []):
    if asset['name'] == '${pattern}':
        print(asset['url'])
        break
" 2>/dev/null || true)"
        if [ -n "$asset_url" ]; then
            echo "${asset_url}|${pattern}"
            return
        fi
    done

    err "No nanobot binary found for platform '${platform}' in release ${version}"
    err "Available assets:"
    echo "$json" | python3 -c "
import sys, json
data = json.load(sys.stdin)
for a in data.get('assets', []):
    print(f'  - {a[\"name\"]}')
" 2>/dev/null
    exit 1
}

main() {
    log "nanobot installer"
    echo ""

    # Resolve version
    if [ "$VERSION" = "latest" ]; then
        VERSION="$(get_latest_version)"
        log "Latest version: ${CYAN}${VERSION}${NC}"
    else
        log "Installing version: ${CYAN}${VERSION}${NC}"
    fi

    # Detect platform
    local platform
    platform="$(detect_platform)"
    log "Platform: ${CYAN}${platform}${NC}"

    # Find download URL
    local result asset_url asset_name
    result="$(find_asset_url "$platform" "$VERSION")"
    asset_url="${result%%|*}"
    asset_name="${result##*|}"

    log "Downloading ${asset_name}..."

    # Create temp dir
    local tmpdir
    tmpdir="$(mktemp -d)"
    trap 'rm -rf "$tmpdir"' EXIT

    # Download
    local archive="${tmpdir}/${asset_name}"
    curl -fsSL -H "Accept: application/octet-stream" -o "$archive" "$asset_url"

    # Create install dir
    mkdir -p "$INSTALL_DIR"

    # Extract
    local binary_name="nanobot"
    if [[ "$platform" == windows-* ]]; then
        binary_name="nanobot.exe"
    fi

    case "$asset_name" in
        *.tar.gz)
            tar xzf "$archive" -C "$tmpdir"
            ;;
        *.zip)
            unzip -o -q "$archive" -d "$tmpdir"
            ;;
    esac

    # Find and move binary
    local found_bin=""
    # Check direct extraction
    if [ -f "${tmpdir}/${binary_name}" ]; then
        found_bin="${tmpdir}/${binary_name}"
    elif [ -f "${tmpdir}/nanobot-bin/${binary_name}" ]; then
        found_bin="${tmpdir}/nanobot-bin/${binary_name}"
    else
        # Search recursively
        found_bin="$(find "$tmpdir" -name "$binary_name" -type f | head -1)"
    fi

    if [ -z "$found_bin" ]; then
        err "Could not find nanobot binary in archive"
        exit 1
    fi

    # Install
    cp -f "$found_bin" "${INSTALL_DIR}/${binary_name}"
    chmod +x "${INSTALL_DIR}/${binary_name}"

    # Verify
    local installed_path="${INSTALL_DIR}/${binary_name}"
    local ver_output
    ver_output="$("$installed_path" --version 2>&1 || echo 'unknown')"

    echo ""
    log "✅ Installed successfully!"
    log "  Binary:  ${installed_path}"
    log "  Version: ${ver_output}"
    log "  Size:    $(ls -lh "$installed_path" | awk '{print $5}')"
    echo ""
    log "hey-nanobot desktop app will auto-detect this binary."
    echo ""

    # Optionally add to PATH (not required for hey-nanobot, but useful for CLI)
    if ! echo "$PATH" | grep -q "$INSTALL_DIR" 2>/dev/null; then
        log "💡 Add to PATH for CLI access:"
        log "   export PATH=\"${INSTALL_DIR}:\$PATH\""
    fi
}

main
