# ScreenshooterMCP

MCP server enabling AI agents to take screenshots and locate UI elements.

## WARNING: IMPORTANT NOTICE

### Git-Worker Requirements

When performing git operations, the following rules **MUST** be followed:

- **NEVER auto-commit** - only commit when explicitly requested by the user
- **ALWAYS use `-s` or `--signoff` flag** for DCO (Developer Certificate of Origin)
- **Title format**: `subsystem: change description` (lowercase, concise)
- **Message**: explain WHY the change was made, not HOW
- **Fixes clause**: when fixing a problem, add `Fixes: <commit hash> (commit title)` between title and body
- **Amending**: do NOT remove sign-off when amending - always use `-s` flag in `git commit --amend`
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
                      │ or HTTP (--listen flag)
┌─────────────────────▼───────────────────────────────┐
│                  MCP Server (Go)                     │
│  ┌─────────────┐  ┌─────────────┐  ┌───────────────┐  │
│  │   tools/    │  │   config/  │  │   capture/    │  │
│  │ capture_*   │──│             │──│   x11/        │  │
│  │ list_*      │  │             │  │   wayland/    │  │
│  └─────────────┘  └─────────────┘  └───────────────┘  │
└─────────────────────────────────────────────────────┘
```

## Key Directories

| Directory | Purpose |
|-----------|---------|
| `cmd/screenshooter-mcp-server/` | Main entrypoint |
| `internal/tools/` | MCP tool implementations |
| `internal/config/` | Configuration loading |
| `internal/capture/` | Common types, interfaces |
| `internal/capture/x11/` | X11 capture implementation |
| `internal/capture/wayland/` | Wayland capture implementation |

## Build & Test

```bash
eval "$(direnv export bash)" && go build ./cmd/screenshooter-mcp-server    # Build
eval "$(direnv export bash)" && go test ./...             # Test all
```

## CLI Options

```bash
screenshot-mcp-server [options]
  -v, --version           Show version
  -h, --help              Show help
  --config                Path to config file
  -l, --log-level         Log level: debug|info|warn|error (default: info)
  --color                 Color output: always|never|auto (default: auto)
  --listen                Listen on TCP address (e.g. 127.0.0.1:8080)
```

## Configuration

Configuration is loaded from (in order of priority):
1. `--config` CLI flag
2. `SCREENSHOOTER_CONFIG` environment variable
3. Default: `~/.local/share/screenshooter-mcp/config.json`

Default config:
```json
{
  "log_level": "info",
  "color": "auto"
}
```

## MCP Tools

| Tool | Description |
|------|-------------|
| `list_monitors` | List available monitors with names and aliases |
| `list_windows` | List open windows with titles and IDs |
| `capture_screen` | Capture full screen or specific monitor - returns PNG image |
| `capture_window` | Capture window by title (partial match supported) - returns PNG image |
| `capture_region` | Capture region from virtual screen - returns PNG image |

## Monitor Naming

Monitors are named using human-readable names with aliases:

```json
{
  "name": "1920x1080-left",
  "aliases": ["DP-1", "monitor-1", "1"],
  "x": 0, "y": 0,
  "width": 1920, "height": 1080
}
```

## Go Development Environment

- **Module**: `github.com/emmanuel-deloget/screenshooter-mcp`
- **Vendoring**: Not used
- **Local GOPATH**: Managed via `.envrc` with direnv
  - Modules cached in `./.go/pkg/mod`
  - Binaries installed to `./.go/bin`

## Environment Auto-Detection

On startup, detect X11 vs Wayland:
1. Check `XDG_SESSION_TYPE` env var
2. Fallback: check for X11 socket (`DISPLAY` set) vs Wayland socket (`WAYLAND_DISPLAY` set)
3. Exit with error if no desktop environment detected

## Distribution

- **Binary**: Just `go build` the server and distribute the single binary
- **No bundled runtime**: Vision API support planned for future (user provides their own)

## Dependencies

| Package | Purpose |
|---------|---------|
| `github.com/jessevdk/go-flags` | CLI argument parsing |
| `github.com/modelcontextprotocol/go-sdk` | MCP protocol |
| `github.com/rs/zerolog` | Structured logging |
| `github.com/nskaggs/perfuncted` | Screen capture (X11, Wayland, Portal) |
| `github.com/jezek/xgb` | X11 bindings for multi-monitor support |

## Testing

- Standard Go `testing` package
- Unit tests in `*_test.go` files
- Run tests: `go test ./...`

## CI/CD

GitHub Actions workflow in `.github/workflows/ci.yml`:
- Build and test on push/PR
- Go vet linting
- Security vulnerability scanning with govulncheck