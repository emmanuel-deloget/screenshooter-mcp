# ScreenshooterMCP

[![CI](https://github.com/emmanuel-deloget/screenshooter-mcp/actions/workflows/ci.yml/badge.svg)](https://github.com/emmanuel-deloget/screenshooter-mcp/actions/workflows/ci.yml)
[![Go Version](https://img.shields.io/badge/Go-1.26-blue)](https://go.dev/)
[![License: MIT](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

MCP server enabling AI agents to take screenshots on Linux (X11 and Wayland).

## Features

### Screen Capture

- **list_monitors** - List available displays with names, positions, and dimensions
- **list_windows** - List open windows with titles, IDs, and state (active, minimized, maximized, fullscreen)
- **capture_screen** - Capture full screen or specific monitor (returns PNG)
- **capture_window** - Capture window by title (partial match supported)
- **capture_region** - Capture rectangular region from screen (returns PNG)

### AI Vision Analysis

- **list_vision_providers** - List configured AI vision providers
- **analyze_image** - Analyze an image with a custom prompt
- **extract_text** - Extract text as formatted markdown (OCR)
- **find_region** - Find bounding box coordinates of a described element
- **compare_images** - Compare two images and describe the differences

### Pipeline Execution

- **execute_capture_pipeline** - Chain multiple capture and vision operations in a single call

### Agent Tools

- **get_skill_info_for_agent** - Return agent skill documentation with tool descriptions and workflow examples
- **list_tool_access** - List all tools with their current access status (allow, deny, ask)
- **allow_tool_access** - Grant temporary access to a restricted tool

## Installation

### From Packages

Pre-compiled packages for Debian/Ubuntu, Fedora, and Arch Linux (x86_64 and ARM64):

| Distribution  | Package Type | Install |
|---------------|-------------|---------|
| Debian/Ubuntu | `.deb` | `dpkg -i screenshooter-mcp-*.deb` |
| Fedora        | `.rpm` | `dnf install screenshooter-mcp-*.rpm` |
| Arch Linux    | `.pkg.tar.zst` | `pacman -U screenshooter-mcp-*.pkg.tar.zst` |

> **Note on ARM64 support**: While x86_64 is fully supported across all distributions, ARM64 (aarch64) packages are provided on a best-effort basis. Arch Linux does not officially support aarch64, so ARM64 packages for this distribution are experimental. Contributions are welcome, but no official support will be provided for ARM64 platforms.

### ⚠️ Security Notice - Automatic Screenshot Authorization

**Server packages automatically pre-authorize screenshot permissions** by configuring the XDG portal permission store. This bypasses the authorization dialog that applications typically receive when requesting screen capture.

This means:
- The MCP server can capture the screen without user prompts
- **All applications** can capture the screen without user prompts (same effect as allowing once)
- On first login, a systemd service runs to grant this permission automatically

This design prioritizes convenience for AI agent use cases but may not be suitable for high-security environments. Future updates may restrict authorization to only the MCP server process.

### GNOME Shell Extension

**Server packages include a GNOME Shell extension** (`screenshooter-mcp@deloget.com`) that provides window management capabilities via D-Bus. This extension is required because modern GNOME Shell versions restrict access to `org.gnome.Shell.Eval()`, which the server previously used to enumerate and manage windows.

The extension exposes the `org.screenshooter.mcp.Windows` D-Bus interface at `/org/screenshooter/mcp`, providing methods for listing windows, activating them, and manipulating their position and size. Two versions are bundled:

| Version | GNOME Shell | API Style |
|---------|-------------|-----------|
| `legacy` | 43, 44 | Imports-based (`imports.gi`) |
| `modern` | 45+ | ES modules (`gi://Gio`) |

On first startup, the systemd service runs `authorize-portal.sh` which automatically detects the GNOME Shell version, copies the appropriate extension to `~/.local/share/gnome-shell/extensions/`, and enables it. The server then queries this D-Bus interface as a fallback when the standard window backend is unavailable.

### Static Binaries

Pre-compiled static binaries are available for all other Linux distributions.

### From Source

```bash
go build -o screenshooter-mcp ./cmd/screenshooter-mcp-server
```

## Usage

### Stdio Mode (Default)

Run without arguments for stdio mode (works with Claude Desktop, Cursor, etc.):

```bash
./screenshooter-mcp
```

### HTTP Server Mode

Run as HTTP server:

```bash
./screenshooter-mcp --listen 127.0.0.1:11777
```

Or configure in config file:

```json
{
  "listen": "127.0.0.1:11777"
}
```

### Configuration

Config file locations (in priority order):

1. `$XDG_CONFIG_HOME/screenshooter-mcp/config.json` (default: `~/.config/screenshooter-mcp/config.json`)
2. `/etc/screenshooter-mcp/config.json` (system default)

Options:
| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `log_level` | string | `"info"` | Log level: debug, info, warn, error |
| `log_format` | string | `"text"` | Log format: text, json |
| `color` | string | `"auto"` | Color output: always, never, auto |
| `listen` | string | `""` | HTTP listen address or empty for stdio |
| `vision` | object | `null` | AI vision providers configuration |
| `access` | object | `{}` | Tool access policies (tool name → "allow", "deny", "ask") |
| `temp_access_duration` | int | `300` | Default duration in seconds for temporary access grants |

### CLI Options

```
  -v, --version                 Show version
  -h, --help                    Show help
      --config=                 Path to config file
  -l, --log-level=              Log level (default: info)
      --log-format=             Log format: text|json (default: text)
      --color=                  Color output: always|never|auto (default: auto)
      --listen=                 Listen on TCP address (e.g. 127.0.0.1:11777) or 'stdio' for stdio mode
      --stdio                   Run in stdio mode (overrides --listen)
      --enable-vision-fallback  Enable automatic fallback to next vision provider on error or timeout
```


## Configuration

Configuration is loaded from (in order of priority):
1. `--config` CLI flag
2. `SCREENSHOOTER_CONFIG` environment variable
3. User config: `$XDG_CONFIG_HOME/screenshooter-mcp/config.json` (default: `~/.config/screenshooter-mcp/config.json`)
4. System config: `/etc/screenshooter-mcp/config.json`

Default config:
```json
{
  "log_level": "info",
  "log_format": "text",
  "color": "auto",
  "listen": "",
  "temp_access_duration": 300
}
```

## MCP Tools

### list_monitors

List all available monitors with their names and aliases.

```json
{
  "Name": "DP-1.1",
  "Aliases": ["right-1920x1080", "DP-1.1", "1"],
  "X": 1920, "Y": 0,
  "Width": 1920, "Height": 1080
}
```

### list_windows

List all open windows with their titles, X11 window IDs, and state (active, minimized, maximized, fullscreen).

### capture_screen

Capture the full screen or a specific monitor.

Arguments:
- `monitor` (optional): Monitor name or alias

### capture_window

Capture a window by its title (partial match supported).

Arguments:
- `title`: Window title to capture

### capture_region

Capture a region from the virtual screen.

Arguments:
- `x`: X coordinate
- `y`: Y coordinate
- `width`: Width
- `height`: Height

### list_vision_providers

List all configured AI vision providers with their names, models, and default status.

### analyze_image

Analyze an image using AI vision providers with a custom prompt.

Arguments:
- `image_base64`: Base64-encoded PNG image data
- `prompt`: Text prompt describing what analysis to perform
- `provider` (optional): Provider name; uses default if not specified
- `timeout` (optional): Timeout in seconds; 0 uses provider default
 
### extract_text

Extract text from an image as formatted markdown (OCR).

Arguments:
- `image_base64`: Base64-encoded PNG image data
- `provider` (optional): Provider name; uses default if not specified
- `timeout` (optional): Timeout in seconds; 0 uses provider default

### find_region

Find bounding box coordinates of a described element in an image.

Arguments:
- `image_base64`: Base64-encoded PNG image data
- `description`: Description of the element to find
- `provider` (optional): Provider name; uses default if not specified
- `timeout` (optional): Timeout in seconds; 0 uses provider default

### compare_images

Compare two images and describe the differences.

Arguments:
- `image_base64`: Base64-encoded PNG image data (first image)
- `image2_base64`: Base64-encoded PNG image data (second image)
- `prompt` (optional): Comparison prompt; uses default if not specified
- `provider` (optional): Provider name; uses default if not specified
- `timeout` (optional): Timeout in seconds; 0 uses provider default

### execute_capture_pipeline

Execute a pipeline of capture and vision operations. Each step's output is pushed onto a stack for use by subsequent steps.

Arguments:
- `pipeline`: Ordered list of steps, each with `tool` and `parameters` fields

Supported pipeline tools: `capture_screen`, `capture_window`, `capture_region`, `find_region`, `extract_text`, `analyze_image`, `compare_images`, `wait_for`.

Example:
```json
{
  "pipeline": [
    {"tool": "capture_window", "parameters": {"title": "Terminal"}},
    {"tool": "find_region", "parameters": {"description": "the error dialog"}},
    {"tool": "capture_region", "parameters": {}},
    {"tool": "extract_text", "parameters": {}}
  ]
}
```

### get_skill_info_for_agent

Return the agent skill documentation. This tool provides a comprehensive guide
to all available tools, workflow examples, and pipeline usage patterns.

## Tool Access Policies

Control which tools are available using the `access` object in `config.json`.
Each tool can be set to `"allow"`, `"deny"`, or `"ask"`.

**Default behavior:**
- `list_monitors`, `list_vision_providers`, `get_skill_info_for_agent`, `list_tool_access`, `allow_tool_access` are always allowed
- All other tools default to `"ask"` if not specified

**Policy effects:**
- `"allow"`: Tool is always available
- `"deny"`: Tool returns an error: `access to tool "..." is denied`
- `"ask"`: Tool returns an error: `access to tool "..." requires user permission; call 'allow_tool_access' to grant temporary access`

**Example config:**
```json
{
  "access": {
    "capture_screen": "deny",
    "capture_window": "ask",
    "capture_region": "allow"
  },
  "temp_access_duration": 300
}
```

### list_tool_access

List all tools with their current access status (`allow`, `deny`, `ask`).

### allow_tool_access

Grant temporary access to a tool that has `"ask"` policy.
Access is granted for a limited time (default: 300 seconds, configurable via `temp_access_duration` in config).

Arguments:
- `tool`: Tool name to grant access to
- `duration` (optional): Duration in seconds (defaults to `temp_access_duration`)

## Vision Providers

Configure AI vision providers in your config file to enable image analysis:

```json
{
  "vision": {
    "enable_fallback": true,
    "providers": [
      {
        "name": "ollama",
        "type": "openai-compatible",
        "base_url": "http://localhost:11434/v1",
        "model": "llava:7b",
        "timeout": 30
      },
      {
        "name": "openai",
        "type": "openai-compatible",
        "model": "gpt-4o",
        "api_key": "sk-...",
        "timeout": 20
      },
      {
        "name": "claude",
        "type": "anthropic",
        "model": "claude-sonnet-4-20250514",
        "api_key": "sk-ant-...",
        "timeout": 20
      },
      {
        "name": "huggingface",
        "type": "huggingface",
        "model": "org/vision-model",
        "api_key": "hf_...",
        "timeout": 30
      },
      {
        "name": "nvidia",
        "type": "openai-compatible",
        "base_url": "https://ai.api.nvidia.com/v1",
        "model": "nvidia/neva-22b",
        "api_key": "nvapi-...",
        "timeout": 30
      },
      {
        "name": "gemini",
        "type": "gemini",
        "model": "gemini-2.5-flash",
        "api_key": "AIza-...",
        "timeout": 30
      },
      {
        "name": "gemini-vertex",
        "type": "gemini",
        "model": "gemini-2.5-flash",
        "project": "my-gcp-project",
        "location": "us-central1",
        "timeout": 30
      }
    ]
  }
}
```

Provider types:
- `openai-compatible`: Works with OpenAI, Ollama, Mistral, Groq, NVIDIA NIM, and any OpenAI-compatible API
- `anthropic`: Anthropic Claude API
- `huggingface`: HuggingFace Inference API
- `gemini`: Google Gemini API or Vertex AI

### Google Gemini

The `gemini` provider supports two modes:

**Gemini API mode** (direct API access):
- Requires `api_key` (get one at [ai.google.dev](https://ai.google.dev/gemini-api/docs/api-key))
- Use model names like `gemini-2.5-flash`, `gemini-2.0-flash`, `gemini-1.5-pro`

**Vertex AI mode** (GCP):
- Requires `project` and `location` (e.g., `us-central1`, `europe-west4`)
- Uses Application Default Credentials for authentication, or set `api_key` for Vertex AI express mode
- Supports all Gemini models available on Vertex AI, plus third-party models like `meta/llama-3.2-90b-vision-instruct`

The first provider in the list is used by default. Specify `provider` in tool calls to use a different one. Timeout is in seconds (default: 20).

### Vision Provider Fallback

When `enable_fallback` is set to `true` (or `--enable-vision-fallback` is passed on the command line), the server automatically tries the next provider in the list if the current one fails or times out. Providers are tried in configuration order until one succeeds or all have been exhausted.

This is useful when your primary provider is unavailable — for example, a local Ollama instance that isn't running, or a cloud API when the network is down.

```json
{
  "vision": {
    "enable_fallback": true,
    "providers": [
      {
        "name": "ollama",
        "type": "openai-compatible",
        "base_url": "http://localhost:11434/v1",
        "model": "qwen3-vl:2b",
        "timeout": 30
      },
      {
        "name": "openai",
        "type": "openai-compatible",
        "model": "gpt-4o",
        "api_key": "sk-...",
        "timeout": 20
      }
    ]
  }
}
```

In this example, if the local Ollama provider fails, the server automatically falls back to OpenAI.

For `compare_images`, providers that do not support image comparison are skipped during fallback.

### Model Limitations

In practice, using small vision models through a local setup (such as llama.cpp or Ollama) may not produce reliable results for all types of vision requests. Notably, the `find_region` tool, which requires the model to identify precise bounding box coordinates, is a particularly difficult task for any vision model — including large models accessed through cloud APIs. Small vision models (such as Qwen3-VL, Moondream, or LLaVA variants) are even more likely to fail or return inaccurate coordinates for this type of request. For best results with `find_region`, consider using a larger, more capable model.

## Requirements

### X11

- X11 server with RANDR extension

### Wayland

- wlroots-based compositor (recommended)
- or xdg-desktop-portal backend

## Testing

Integration tests create VMs using KVM/libvirt to test the MCP server end-to-end across supported desktop environments. Each test provisions a VM, installs the server, and runs MCP tool calls (`list_monitors`, `list_windows`, `capture_screen`, `capture_region`) against it.

To run a single test:
```bash
cd tests/integration
./run.sh debian 12 gnome wayland
```

To run all tests:
```bash
cd tests/integration
./all.sh
```

See `tests/integration/README.md` for requirements and supported configurations.

### Test Results

| Distribution / Version | Desktop / Mode | Status | Notes |
|------------------------|----------------|--------|-------|
| Debian 12 | GNOME / Wayland | ✅  | ⚠️ Uses `screenshooter-mcp@deloget.com` GNOME extension |
| Debian 12 | GNOME / X11 | ✅  |  |
| Debian 12 | KDE / Wayland | ✅  | |
| Debian 12 | KDE / X11 | ✅  |  |
| Debian 13 | GNOME / Wayland | ✅  | ⚠️ Uses `screenshooter-mcp@deloget.com` GNOME extension |
| Debian 13 | GNOME / X11 | ✅  |  |
| Debian 13 | KDE / Wayland | ✅  | |
| Debian 13 | KDE / X11 | ✅  |  |
| Fedora 43 | GNOME / Wayland | ✅  | ⚠️ Uses `screenshooter-mcp@deloget.com` GNOME extension |
| Fedora 43 | KDE / Wayland | ✅  | |
| Ubuntu 24.04 | GNOME / Wayland | ✅  | ⚠️ Uses `screenshooter-mcp@deloget.com` GNOME extension |
| Ubuntu 24.04 | GNOME / X11 | ✅  |  |
| Ubuntu 24.04 | KDE / Wayland | ✅  |  |
| Ubuntu 24.04 | KDE / X11 | ✅  |  |
| Ubuntu 25.10 | GNOME / Wayland | ✅  | ⚠️ Uses `screenshooter-mcp@deloget.com` GNOME extension |
| Ubuntu 25.10 | KDE / Wayland | ✅  |  |
| Arch Linux latest | GNOME / Wayland | ✅  | ⚠️ Uses `screenshooter-mcp@deloget.com` GNOME extension |
| Arch Linux latest | KDE / Wayland | ✅  |  |

## License

MIT