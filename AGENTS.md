# ScreenshooterMCP

MCP server enabling AI agents to take screenshots and locate UI elements using a local vision model.

## WARNING: IMPORTANT NOTICE

### Git-Worker Requirements

When performing git operations, the following rules **MUST** be followed:

- **NEVER auto-commit** - only commit when explicitly requested by the user
- **ALWAYS use `-s` or `--signoff` flag** for DCO (Developer Certificate of Origin)
- **Title format**: `subsystem: change description` (lowercase, concise)
- **Message**: explain WHY the change was made, not HOW
- **Fixes clause**: when fixing a problem, add `Fixes: <commit hash> (commit title)` between title and body
- **ALL config changes** MUST use `--local` flag: `git config --local ...`
- When multiple commits are needed, **SHOW THE PLAN** before proceeding

### Commit Workflow

1. Run `git status`, `git diff`, and `git log` to understand current state
2. Draft commit message: title + body explaining WHY
3. Stage with `git add <files>`
4. Commit with `git commit -s -m "title\n\nbody"`

## Architecture

```
┌─────────────────────────────────────────────────────┐
│                    MCP Client                        │
│               (Claude Desktop, etc.)                │
└─────────────────────┬───────────────────────────────┘
                      │ stdio (MCP over stdin/stdout)
┌─────────────────────▼───────────────────────────────┐
│                  MCP Server (Go)                     │
│  ┌─────────────┐  ┌─────────────┐  ┌───────────────┐  │
│  │   tools/    │  │   vision/   │  │   capture/    │  │
│  │  screenshot │──│  ollama.go  │──│   platform/   │  │
│  │  find_elem  │  │   client    │  │  (X11/Wayland)│  │
│  └─────────────┘  └─────────────┘  └───────────────┘  │
└─────────────────────┬───────────────────────────────┘
                      │ HTTP
┌─────────────────────▼───────────────────────────────┐
│              Ollama (embedded in AppImage)           │
│    Local VLM (moondream2, llava-llama3, etc.)      │
└─────────────────────────────────────────────────────┘
```

## Key Directories

| Directory | Purpose |
|-----------|---------|
| `cmd/server/` | Main entrypoint, MCP transport setup |
| `internal/tools/` | MCP tool implementations |
| `internal/vision/` | Ollama API client, prompt templates |
| `internal/capture/` | Platform-specific screen capture |
| `internal/capture/x11/` | X11 capture via xlib or xcb |
| `internal/capture/wayland/` | Wayland capture via xdg-desktop-portal |
| `assets/models/` | Bundled model manifests |

## Build & Test

```bash
go build ./cmd/server          # Build
go test ./...                  # Test all packages
go build -tags=x11 ./cmd/server   # Build with X11 support
go build -tags=wayland ./cmd/server # Build with Wayland support
```

## MCP Tools

- `take_screenshot` - Capture full screen or specific window (params: `display?`, `window_id?`)
- `find_element` - Locate UI element in screenshot (params: `image`, `description`) → returns bounding box `[x1,y1,x2,y2]`
- `click_at` - (future) Simulate click at coordinates

## Ollama Integration

- Embedded Ollama binary inside AppImage
- Default endpoint: `http://127.0.0.1:11434`
- Model configured via `OLLAMA_MODEL` env var (default: `moondream2`)
- Prompt template outputs bounding box coordinates

## Environment Auto-Detection

On startup, detect X11 vs Wayland:
1. Check `XDG_SESSION_TYPE` env var
2. Fallback: check for X11 socket (`DISPLAY` set) vs Wayland socket (`WAYLAND_DISPLAY` set)
3. If capture API unavailable, log warning with missing package instructions and exit gracefully

### Required Packages (auto-detected warnings)

| Environment | Required | Package (Debian/Ubuntu) |
|-------------|----------|-------------------------|
| X11 | libx11-dev, libxcb1-dev | `apt install libx11-dev libxcb1-dev` |
| Wayland | libwayland-dev, xdg-desktop-portal | `apt install libwayland-dev` |
| Both | Ollama running | `curl http://127.0.0.1:11434` |

## Vision Model

- **CPU-only**: Designed to run efficiently without GPU
- **Recommended**: `moondream2` (~1.4B params, fast on CPU)
- **Alternatives**: `llava-llama3`, `qwen2-vl`
- **Storage**: `~/.local/share/screenshooter-mcp/models/` (primary) or `~/.ollama/models/` (fallback)
- **First-run**: Auto-download model if not cached; skip if already present

## Distribution

- **AppImage**: Bundles Go MCP server + Ollama binary (~100-200MB)
- **First-run**: Check for cached model; download if missing (one-time ~2-4GB)
- **Model cache**: `~/.local/share/screenshooter-mcp/models/` persists across updates

## Security Notes

- Screen capture requires elevated access to display server
- MCP server validates all tool arguments
- Ollama runs locally (no data leaves the machine)
