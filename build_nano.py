#!/usr/bin/env python3
"""
PyInstaller build script for nanobot-ai.

Bundles nanobot into a standalone single-file binary using the uv-managed
Python environment that has all dependencies installed.

Usage:
    cd /path/to/nanobot-desktop
    python3 build_nano.py

Output:
    nanobot-bin/nanobot          (macOS / Linux)
    nanobot-bin\\nanobot.exe      (Windows)

Requirements:
    pip install pyinstaller
"""

from __future__ import annotations

import os
import platform
import shutil
import subprocess
import sys
from pathlib import Path

# ---------------------------------------------------------------------------
# Configuration
# ---------------------------------------------------------------------------

# The uv tool install root — adjust if your setup differs.
# Override with env var NANOBOT_ROOT to support CI environments.
_env_root = os.environ.get("NANOBOT_ROOT")
if _env_root:
    UV_TOOL_ROOT = Path(_env_root)
else:
    UV_TOOL_ROOT = Path.home() / ".local/share/uv/tools/nanobot-ai"

# Override site-packages with env var (for venv-based CI installs)
_env_site = os.environ.get("NANOBOT_SITE_PACKAGES")
if _env_site:
    SITE_PACKAGES = Path(_env_site)
else:
    SITE_PACKAGES = UV_TOOL_ROOT / "lib/python3.13/site-packages"

NANOBOT_PKG = SITE_PACKAGES / "nanobot"

# Python interpreter — override with NANOBOT_PYTHON for CI
_env_python = os.environ.get("NANOBOT_PYTHON")
if _env_python:
    UV_PYTHON = Path(_env_python)
else:
    UV_PYTHON = UV_TOOL_ROOT / "bin" / "python3"

# Output directory (relative to this script's location)
SCRIPT_DIR = Path(__file__).resolve().parent
OUTPUT_DIR = SCRIPT_DIR / "nanobot-bin"
BUILD_DIR = SCRIPT_DIR / "build_nanobot"
SPEC_FILE = SCRIPT_DIR / "build_nano.spec"

# Binary name
SYSTEM = platform.system()
BINARY_NAME = "nanobot.exe" if SYSTEM == "Windows" else "nanobot"

# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------

def banner(msg: str) -> None:
    width = 60
    print()
    print("=" * width)
    print(f"  {msg}")
    print("=" * width)


def info(msg: str) -> None:
    print(f"  [INFO] {msg}")


def warn(msg: str) -> None:
    print(f"  [WARN] {msg}")


def error(msg: str) -> None:
    print(f"  [ERROR] {msg}")


def run(cmd: list[str], **kwargs) -> subprocess.CompletedProcess:
    """Run a command with real-time output."""
    info(f"Running: {' '.join(str(c) for c in cmd)}")
    result = subprocess.run(cmd, **kwargs)
    if result.returncode != 0:
        error(f"Command failed with exit code {result.returncode}")
        if result.stdout:
            print(result.stdout.decode() if isinstance(result.stdout, bytes) else result.stdout)
        if result.stderr:
            print(result.stderr.decode() if isinstance(result.stderr, bytes) else result.stderr)
    return result


# ---------------------------------------------------------------------------
# Data collection
# ---------------------------------------------------------------------------

def collect_data_files() -> list[tuple[str, str]]:
    """Collect non-Python data files that nanobot needs at runtime.

    Returns list of (source_path, dest_subdir) tuples relative to the nanobot package.
    """
    data_files: list[tuple[str, str]] = []

    # Templates (SOUL.md, TOOLS.md, AGENTS.md, USER.md, HEARTBEAT.md, memory/*)
    templates_dir = NANOBOT_PKG / "templates"
    if templates_dir.exists():
        for f in templates_dir.rglob("*"):
            if f.is_file() and f.suffix != ".py":
                rel = f.relative_to(NANOBOT_PKG)
                data_files.append((str(f), str(rel.parent)))
                info(f"  Data: {rel}")

    # Skills (SKILL.md files, scripts, README.md)
    skills_dir = NANOBOT_PKG / "skills"
    if skills_dir.exists():
        for f in skills_dir.rglob("*"):
            if f.is_file() and f.suffix != ".py":
                rel = f.relative_to(NANOBOT_PKG)
                data_files.append((str(f), str(rel.parent)))
                info(f"  Data: {rel}")

    # Bridge source files (TypeScript — needed for WhatsApp bridge install)
    bridge_dir = NANOBOT_PKG / "bridge"
    if bridge_dir.exists():
        for f in bridge_dir.rglob("*"):
            if f.is_file():
                rel = f.relative_to(NANOBOT_PKG)
                data_files.append((str(f), str(rel.parent)))
                info(f"  Data: {rel}")

    return data_files


def collect_hidden_imports() -> list[str]:
    """Return all hidden imports PyInstaller can't auto-detect.

    Covers:
    - Channel plugins (dynamically imported via importlib)
    - Provider backends
    - Tool modules
    - Third-party packages with dynamic imports
    """
    imports = [
        # ---- nanobot sub-packages (ensure full tree is included) ----
        "nanobot",
        "nanobot.__main__",
        "nanobot.cli",
        "nanobot.cli.commands",
        "nanobot.cli.models",
        "nanobot.cli.onboard",
        "nanobot.cli.stream",
        "nanobot.config",
        "nanobot.config.loader",
        "nanobot.config.paths",
        "nanobot.config.schema",
        "nanobot.agent",
        "nanobot.agent.context",
        "nanobot.agent.hook",
        "nanobot.agent.loop",
        "nanobot.agent.memory",
        "nanobot.agent.runner",
        "nanobot.agent.skills",
        "nanobot.agent.subagent",
        "nanobot.agent.tools",
        "nanobot.agent.tools.base",
        "nanobot.agent.tools.cron",
        "nanobot.agent.tools.filesystem",
        "nanobot.agent.tools.mcp",
        "nanobot.agent.tools.message",
        "nanobot.agent.tools.registry",
        "nanobot.agent.tools.shell",
        "nanobot.agent.tools.spawn",
        "nanobot.agent.tools.web",
        "nanobot.channels",
        "nanobot.channels.base",
        "nanobot.channels.manager",
        "nanobot.channels.registry",
        "nanobot.channels.telegram",
        "nanobot.channels.discord",
        "nanobot.channels.slack",
        "nanobot.channels.qq",
        "nanobot.channels.feishu",
        "nanobot.channels.dingtalk",
        "nanobot.channels.email",
        "nanobot.channels.whatsapp",
        "nanobot.channels.weixin",
        "nanobot.channels.wecom",
        "nanobot.channels.mochat",
        "nanobot.channels.matrix",
        "nanobot.providers",
        "nanobot.providers.base",
        "nanobot.providers.registry",
        "nanobot.providers.anthropic_provider",
        "nanobot.providers.openai_compat_provider",
        "nanobot.providers.openai_codex_provider",
        "nanobot.providers.azure_openai_provider",
        "nanobot.providers.transcription",
        "nanobot.bus",
        "nanobot.bus.events",
        "nanobot.bus.queue",
        "nanobot.command",
        "nanobot.command.builtin",
        "nanobot.command.router",
        "nanobot.cron",
        "nanobot.cron.service",
        "nanobot.cron.types",
        "nanobot.heartbeat",
        "nanobot.security",
        "nanobot.security.network",
        "nanobot.session",
        "nanobot.session.manager",
        "nanobot.utils",
        "nanobot.utils.evaluator",
        "nanobot.utils.helpers",
        "nanobot.skills",
        "nanobot.skills.memory",
        "nanobot.skills.summarize",
        "nanobot.skills.clawhub",
        "nanobot.skills.skill_creator",
        "nanobot.skills.github",
        "nanobot.skills.tmux",
        "nanobot.skills.weather",
        "nanobot.skills.cron",

        # ---- Third-party packages (dynamic / lazy imports) ----
        # Core
        "typer",
        "click",                    # typer wraps click
        "rich",
        "rich.console",
        "rich.markdown",
        "rich.table",
        "rich.text",
        "rich.syntax",
        "rich.theme",
        "loguru",
        "pydantic",
        "pydantic.alias_generators",
        "httpx",
        "httpx._transports",
        "httpx._transports.default",
        "openai",
        "anthropic",

        # CLI / prompts
        "prompt_toolkit",
        "prompt_toolkit.application",
        "prompt_toolkit.formatted_text",
        "prompt_toolkit.history",
        "prompt_toolkit.patch_stdout",

        # Channels
        "telegram",
        "telegram.ext",
        "telegram.request",
        "slack_sdk",
        "slack_sdk.socket_mode",
        "slack_sdk.socket_mode.websockets",
        "slack_sdk.web.async_client",
        "dingtalk_stream",
        "dingtalk_stream.chatbot",
        "lark_oapi",
        "aiohttp",
        "websockets",
        "slackify_markdown",
        "nio",                      # Matrix (optional — may not be installed)

        # Search / tools
        "ddgs",
        "mcp",
        "croniter",

        # Misc
        "yaml",
        "json_repair",
        "jsonschema",
        "dotenv",
        "chardet",
        "oauth_cli_kit",
        "uvicorn",
        "markdown_it",
        "pygments",
        "tqdm",
        "distro",
        "normalizer",
    ]
    return imports


# ---------------------------------------------------------------------------
# PyInstaller spec generation
# ---------------------------------------------------------------------------

def generate_spec() -> str:
    """Generate a PyInstaller .spec file for nanobot."""
    data_files = collect_data_files()
    hidden_imports = collect_hidden_imports()

    # Format hidden imports
    hi_lines = ",\n        ".join(f'"{imp}"' for imp in hidden_imports)

    # Format data files as (dest, [source]) tuples
    df_lines = ",\n        ".join(
        f'("{dest}", ["{src}"])' for src, dest in data_files
    )

    spec_content = f'''# -*- mode: python ; coding: utf-8 -*-
"""
PyInstaller spec for nanobot — auto-generated by build_nano.py.

To build manually:
    cd /path/to/nanobot-desktop
    pyinstaller build_nano.spec
"""

import sys
from pathlib import Path

block_cipher = None

# Resolve the uv site-packages so the spec works outside that env.
SITE_PKGS = r"{SITE_PACKAGES}"
NANOBOT_PKG = Path(SITE_PKGS) / "nanobot"

a = Analysis(
    # Entry point: a thin wrapper that calls typer app()
    [r"{SCRIPT_DIR / "_nanobot_entry.py"}"],
    pathex=[
        r"{SITE_PACKAGES}",
    ],
    binaries=[],
    datas=[
        # nanobot package data files
        {df_lines}
    ],
    hiddenimports=[
        {hi_lines}
    ],
    hookspath=[],
    hooksconfig={{}},
    runtime_hooks=[],
    excludes=[
        # Trim heavy optional packages we don't need
        "tkinter",
        "unittest",
        "test",
        "tests",
        "setuptools",
        "pip",
    ],
    win_no_prefer_redirects=False,
    win_private_assemblies=False,
    cipher=block_cipher,
    noarchive=False,
)

pyz = PYZ(a.pure, a.zipped_data, cipher=block_cipher)

exe = EXE(
    pyz,
    a.scripts,
    a.binaries,
    a.zipfiles,
    a.datas,
    [],
    name="{BINARY_NAME}",
    debug=False,
    bootloader_ignore_signals=False,
    strip=False,
    upx=True,
    upx_exclude=[],
    runtime_tmpdir=None,
    console=True,
    disable_windowed_traceback=False,
    argv_emulation=False,
    target_arch=None,
    codesign_identity=None,
    entitlements_file=None,
    icon=None,
)
'''
    return spec_content


# ---------------------------------------------------------------------------
# Entry point wrapper
# ---------------------------------------------------------------------------

ENTRY_SCRIPT = f'''#!/usr/bin/env python3
"""Nanobot entry point for PyInstaller bundling."""
import sys
import os

# Ensure UTF-8 on Windows
if sys.platform == "win32":
    os.environ.setdefault("PYTHONIOENCODING", "utf-8")

from nanobot.cli.commands import app

if __name__ == "__main__":
    sys.exit(app())
'''


# ---------------------------------------------------------------------------
# Main build
# ---------------------------------------------------------------------------

def main() -> int:
    banner("nanobot PyInstaller Build")

    # 1. Validate environment
    info(f"System: {SYSTEM} ({platform.machine()})")
    info(f"Python: {sys.version}")
    info(f"nanobot package: {NANOBOT_PKG}")
    info(f"Output dir: {OUTPUT_DIR}")
    info(f"Build dir: {BUILD_DIR}")

    if not NANOBOT_PKG.exists():
        error(f"nanobot package not found at {NANOBOT_PKG}")
        error("Make sure nanobot-ai is installed via uv: uv tool install nanobot-ai")
        return 1

    # Find PyInstaller — prefer the configured python, fall back to system
    uv_python = UV_PYTHON
    if not uv_python.exists():
        uv_python = UV_TOOL_ROOT / "bin" / "python3"
    if not uv_python.exists():
        uv_python = UV_TOOL_ROOT / "bin" / "python"
    if SYSTEM == "Windows" and not uv_python.exists():
        uv_python = UV_TOOL_ROOT / "Scripts" / "python.exe"

    # Check if pyinstaller is available
    pyinstaller_cmd = None
    for candidate in [
        UV_TOOL_ROOT / "bin" / "pyinstaller",
        shutil.which("pyinstaller"),
    ]:
        if candidate and Path(candidate).exists():
            pyinstaller_cmd = str(candidate)
            break

    if not pyinstaller_cmd:
        # Try via the python
        if uv_python.exists():
            check = subprocess.run(
                [str(uv_python), "-m", "PyInstaller", "--version"],
                capture_output=True, text=True,
            )
            if check.returncode == 0:
                pyinstaller_cmd = str(uv_python)
                info(f"Will use: {uv_python} -m PyInstaller")
        else:
            error("PyInstaller not found!")
            error("Install it with: uv pip install pyinstaller")
            error("  or: pip install pyinstaller")
            return 1

    info(f"PyInstaller: {pyinstaller_cmd}")

    # 2. Clean previous build artifacts
    if BUILD_DIR.exists():
        info(f"Cleaning build dir: {BUILD_DIR}")
        shutil.rmtree(BUILD_DIR, ignore_errors=True)
    BUILD_DIR.mkdir(parents=True, exist_ok=True)

    # 3. Write the entry point wrapper
    entry_path = BUILD_DIR / "_nanobot_entry.py"
    entry_path.write_text(ENTRY_SCRIPT)
    info(f"Entry script: {entry_path}")

    # 4. Generate the spec file
    spec_content = generate_spec()
    SPEC_FILE.write_text(spec_content)
    info(f"Spec file: {SPEC_FILE}")

    # 5. Build the PyInstaller command
    hidden_imports = collect_hidden_imports()
    data_files = collect_data_files()

    cmd = [
        str(uv_python) if uv_python.exists() else sys.executable,
        "-m", "PyInstaller",
        "--name", BINARY_NAME,
        "--onefile",
        "--clean",
        "--noconfirm",
        "--distpath", str(OUTPUT_DIR),
        "--workpath", str(BUILD_DIR / "work"),
        "--specpath", str(BUILD_DIR),
        f"--paths={SITE_PACKAGES}",
    ]

    # Add hidden imports
    for imp in hidden_imports:
        cmd.extend(["--hidden-import", imp])

    # Add data files
    for src, dest in data_files:
        cmd.extend(["--add-data", f"{src}{os.pathsep}{dest}"])

    # Exclude heavy unneeded packages
    for exc in ["tkinter", "unittest", "test", "tests", "setuptools", "pip"]:
        cmd.extend(["--exclude-module", exc])

    # Console mode
    cmd.append("--console")

    # Entry point
    cmd.append(str(entry_path))

    banner("Running PyInstaller")
    info(f"Hidden imports: {len(hidden_imports)}")
    info(f"Data files: {len(data_files)}")

    result = run(cmd, cwd=str(SCRIPT_DIR))

    if result.returncode != 0:
        error("Build failed!")
        return 1

    # 6. Verify output
    banner("Build Complete")
    output_binary = OUTPUT_DIR / BINARY_NAME

    if not output_binary.exists():
        error(f"Output binary not found: {output_binary}")
        # Check if it ended up somewhere else
        for candidate in OUTPUT_DIR.rglob("nanobot*"):
            info(f"  Found: {candidate}")
        return 1

    # Make executable
    output_binary.chmod(0o755)

    # File size
    size_mb = output_binary.stat().st_size / (1024 * 1024)
    info(f"Binary: {output_binary}")
    info(f"Size: {size_mb:.1f} MB")

    # 7. Quick test
    banner("Testing Binary")
    test_cmd = [str(output_binary), "--version"]
    info(f"Running: {' '.join(test_cmd)}")
    test_result = subprocess.run(test_cmd, capture_output=True, text=True, timeout=15)

    if test_result.returncode == 0:
        info(f"Output: {test_result.stdout.strip()}")
        info("✓ Binary works!")
    else:
        warn(f"Exit code: {test_result.returncode}")
        if test_result.stdout:
            info(f"stdout: {test_result.stdout.strip()}")
        if test_result.stderr:
            warn(f"stderr: {test_result.stderr.strip()}")
        warn("Binary test had issues (may still work for normal usage)")

    banner("Done!")
    return 0


if __name__ == "__main__":
    sys.exit(main())
