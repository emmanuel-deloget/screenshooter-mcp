# ScreenshooterMCP

[![CI](https://github.com/emmanuel-deloget/screenshooter-mcp/actions/workflows/ci.yml/badge.svg)](https://github.com/emmanuel-deloget/screenshooter-mcp/actions/workflows/ci.yml)
[![Go Version](https://img.shields.io/badge/Go-1.26-blue)](https://go.dev/)
[![License: MIT](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

MCP server enabling AI agents to take screenshots on Linux (X11 and Wayland).

## Features

- **list_monitors** - List available displays with names, positions, and dimensions
- **list_windows** - List open windows with titles and IDs
- **capture_screen** - Capture full screen or specific monitor (returns PNG)
- **capture_window** - Capture window by title (partial match supported)
- **capture_region** - Capture rectangular region from screen (returns PNG)

## Installation

### From Packages

Pre-compiled packages for Debian/Ubuntu and Fedora (x86_64 and ARM64):

| Distribution  | Package Type | Install |
|---------------|-------------|---------|
| Debian/Ubuntu | `.deb` | `dpkg -i screenshooter-mcp-*.deb` |
| Fedora        | `.rpm` | `dnf install screenshooter-mcp-*.rpm` |

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
| `color` | string | `"auto"` | Color output: always, never, auto |
| `listen` | string | `""` | HTTP listen address or empty for stdio |

### CLI Options

```
-v, --version           Show version
-h, --help            Show help
--config              Path to config file
-l, --log-level       Log level: debug|info|warn|error
--color               Color output: always|never|auto
--listen             Listen on TCP address (e.g. 127.0.0.1:11777) or 'stdio'
--stdio              Force stdio mode (overrides --listen)
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
  "color": "auto",
  "listen": ""
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

List all open windows with their titles and X11 window IDs.

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
| Ubuntu 24.04 | KDE / Wayland | ❌ | `list_windows` times out — KWin 5 (Plasma 5) uses `clientList()` API, upstream `perfuncted` library only supports KWin 6 `windowList()` |
| Ubuntu 24.04 | KDE / X11 | ✅  |  |
| Ubuntu 25.10 | GNOME / Wayland | ✅  | ⚠️ Uses `screenshooter-mcp@deloget.com` GNOME extension |
| Ubuntu 25.10 | KDE / Wayland | ✅  |  |

## License

MIT