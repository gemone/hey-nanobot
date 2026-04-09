# install.ps1 — Install nanobot binary from GitHub Release (Windows)
#
# Usage:
#   irm https://raw.githubusercontent.com/gemone/hey-nanobot/main/install.ps1 | iex
#   .\install.ps1                    # Install latest
#   .\install.ps1 -Version v1.2.3   # Install specific version
#
# Installs nanobot binary to %USERPROFILE%\.local\share\hey-nanobot\bin\

param(
    [string]$Version = "latest"
)

$ErrorActionPreference = "Stop"
$Repo = "gemone/hey-nanobot"
$InstallDir = if ($env:HEY_NANOBOT_BIN) { $env:HEY_NANOBOT_BIN } else { Join-Path $env:USERPROFILE ".local\share\hey-nanobot\bin" }

function Write-Status($msg) { Write-Host "[install] $msg" -ForegroundColor Green }
function Write-Warn($msg) { Write-Host "[warn] $msg" -ForegroundColor Yellow }
function Write-Err($msg) { Write-Host "[error] $msg" -ForegroundColor Red }

# Detect platform
$Arch = if ([System.Runtime.InteropServices.RuntimeInformation]::ProcessArchitecture -match "ARM64") { "arm64" } else { "amd64" }
Write-Status "Platform: windows-$Arch"

# Resolve version
if ($Version -eq "latest") {
    $releaseUrl = "https://api.github.com/repos/$Repo/releases/latest"
} else {
    $releaseUrl = "https://api.github.com/repos/$Repo/releases/tags/$Version"
}

Write-Status "Fetching release info..."
$release = Invoke-RestMethod -Uri $releaseUrl -Headers @{ "User-Agent" = "hey-nanobot-installer" }
$Version = $release.tag_name
Write-Status "Version: $Version"

# Find asset
$assetPattern = "nanobot-windows-$Arch.zip"
$asset = $release.assets | Where-Object { $_.name -eq $assetPattern } | Select-Object -First 1

if (-not $asset) {
    Write-Err "No nanobot binary found for windows-$Arch in release $Version"
    Write-Err "Available assets:"
    $release.assets | ForEach-Object { Write-Err "  - $($_.name)" }
    exit 1
}

# Download
$tmpDir = [System.IO.Path]::GetTempPath() + [System.Guid]::NewGuid().ToString()
New-Item -ItemType Directory -Path $tmpDir | Out-Null
try {
    $archive = Join-Path $tmpDir $asset.name
    Write-Status "Downloading $($asset.name)..."
    Invoke-WebRequest -Uri $asset.url -Headers @{ "Accept" = "application/octet-stream"; "User-Agent" = "hey-nanobot-installer" } -OutFile $archive

    # Extract
    Write-Status "Extracting..."
    Expand-Archive -Path $archive -DestinationPath $tmpDir -Force

    # Find binary
    $binary = Get-ChildItem -Path $tmpDir -Recurse -Filter "nanobot.exe" | Select-Object -First 1
    if (-not $binary) {
        Write-Err "Could not find nanobot.exe in archive"
        exit 1
    }

    # Install
    New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
    $destPath = Join-Path $InstallDir "nanobot.exe"
    Copy-Item -Path $binary.FullName -Destination $destPath -Force

    # Verify
    $verOutput = & $destPath --version 2>&1
    Write-Host ""
    Write-Status "Installed successfully!"
    Write-Status "  Binary:  $destPath"
    Write-Status "  Version: $verOutput"
    Write-Status "  Size:    $((Get-Item $destPath).Length / 1MB).ToString('F1') MB"
    Write-Host ""
    Write-Status "hey-nanobot desktop app will auto-detect this binary."

} finally {
    Remove-Item -Path $tmpDir -Recurse -Force -ErrorAction SilentlyContinue
}
