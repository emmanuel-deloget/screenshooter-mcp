# ScreenshooterMCP

MCP server enabling AI agents to take screenshots on Linux (X11 and Wayland).

## Features

- **list_monitors** - List available monitors with names and aliases
- **list_windows** - List open windows with titles and IDs
- **capture_screen** - Capture full screen or specific monitor (returns PNG)
- **capture_window** - Capture window by title (partial match supported)
- **capture_region** - Capture region from virtual screen (returns PNG)

## Installation

### From Packages

Available for Debian, Fedora, Arch Linux, and Alpine:

| Distribution | Package Type | Install |
|-------------|-------------|---------|
| Debian/Ubuntu | `.deb` | `dpkg -i screenshooter-mcp-*.deb` |
| Fedora | `.rpm` | `dnf install screenshooter-mcp-*.rpm` |
| Arch Linux | `.pkg.tar.zst` | `pacman -U screenshooter-mcp-*.pkg.tar.zst` |
| Alpine | `.apk` | `apk add screenshooter-mcp-*.apk` |

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
- `perfuncted` library

### Wayland

- wlroots-based compositor (recommended)
- or portal backend (xdg-desktop-portal)

## License

MIT