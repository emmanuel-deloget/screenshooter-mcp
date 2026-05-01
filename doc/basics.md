# Developer Basics

## Project Structure

```
screenshooter-mcp/
├── cmd/screenshooter-mcp-server/   # Main entrypoint, tool registration
├── internal/
│   ├── capture/                    # Screen capture implementations
│   │   ├── types.go                # Core interfaces (ScreenCapture, Window, Monitor)
│   │   ├── element.go              # BoundingBox/Element types for region handling
│   │   ├── environment.go          # Desktop environment detection (X11 vs Wayland)
│   │   ├── x11/                    # X11 capture (RANDR + perfuncted)
│   │   └── wayland/                # Wayland capture (portal + perfuncted + GNOME extension)
│   │       ├── capture.go          # Wayland ScreenCapture implementation
│   │       └── gnome.go            # D-Bus GNOME window manager
│   ├── config/                     # Configuration loading (XDG-based)
│   ├── logging/                    # Structured logging (zerolog)
│   ├── tools/                      # MCP tool implementations
│   │   ├── tools.go                # Core tools (capture_*, list_*, vision_*)
│   │   ├── pipeline.go             # Pipeline executor (stack-based chaining)
│   │   └── skills/
│   │       └── SKILL.md            # Agent usage guide (embedded in binary)
│   └── vision/                     # AI vision providers
│       ├── vision.go               # Manager + Provider interface
│       ├── openai.go               # OpenAI-compatible provider
│       ├── anthropic.go            # Anthropic Claude provider
│       └── huggingface.go          # HuggingFace Inference API provider
├── gnome-extension/                # GNOME Shell extension (D-Bus window management)
│   ├── screenshooter-mcp@deloget.com_legacy/   # GNOME 43/44
│   └── screenshooter-mcp@deloget.com_modern/   # GNOME 45+
├── scripts/packaging/              # Packaging scripts (systemd units, desktop files, etc.)
├── tests/integration/              # End-to-end integration tests (VM-based)
└── .github/workflows/              # CI/CD pipelines
    ├── ci.yml                      # Build, test, vet on PR/push
    └── packages.yml                # Release package builds (deb, rpm, static)
```

## Architecture

```
┌─────────────────────────────────────────────────────┐
│                    MCP Client                       │
│               (Claude Desktop, Cursor, etc.)        │
└─────────────────────┬───────────────────────────────┘
                      │ stdio (MCP over stdin/stdout)
                      │ or HTTP (--listen flag)
┌─────────────────────▼────────────────────────────────┐
│                  MCP Server (Go)                     │
│  ┌─────────────┐  ┌─────────────┐  ┌───────────────┐ │
│  │   tools/    │  │   config/   │  │   capture/    │ │
│  │ capture_*   │──│             │──│   x11/        │ │
│  │ list_*      │  │             │  │   wayland/    │ │
│  │ pipeline    │  │             │  └───────────────┘ │
│  │ vision_*    │  │             │                    │
│  └─────────────┘  └─────────────┘  ┌───────────────┐ │
│                                    │   vision/     │ │
│                                    │   openai      │ │
│                                    │   anthropic   │ │
│                                    │   huggingface │ │
│                                    └───────────────┘ │
└──────────────────────────────────────────────────────┘
```

## Available Tools

| Category | Tool | Description |
|----------|------|-------------|
| Capture | `list_monitors` | List displays with names, positions, dimensions |
| Capture | `list_windows` | List open windows with titles, IDs, and state |
| Capture | `capture_screen` | Capture full screen or specific monitor |
| Capture | `capture_window` | Capture window by title (partial match) |
| Capture | `capture_region` | Capture rectangular region |
| Vision | `list_vision_providers` | List configured AI vision providers |
| Vision | `analyze_image` | Analyze image with custom prompt |
| Vision | `extract_text` | Extract text as markdown (OCR) |
| Vision | `find_region` | Find element bounding box coordinates |
| Vision | `compare_images` | Compare two images, describe differences |
| Pipeline | `execute_capture_pipeline` | Chain capture/vision operations |
| Agent | `get_skill_info_for_agent` | Return agent skill documentation |

## Build & Test

```bash
# Build
eval "$(direnv export bash)" && go build ./cmd/screenshooter-mcp-server

# Test all
eval "$(direnv export bash)" && go test ./...

# Lint
eval "$(direnv export bash)" && go vet ./...

# Format
go fmt ./...
```

## Dependencies

| Package | Purpose |
|---------|---------|
| `github.com/jessevdk/go-flags` | CLI argument parsing |
| `github.com/modelcontextprotocol/go-sdk` | MCP protocol |
| `github.com/rs/zerolog` | Structured logging |
| `github.com/nskaggs/perfuncted` | Screen capture (X11, Wayland, Portal) |
| `github.com/jezek/xgb` | X11 bindings for multi-monitor support (RANDR) |
| `github.com/godbus/dbus` | D-Bus communication (GNOME extension) |
| `github.com/sashabaranov/go-openai` | OpenAI-compatible API client |
| `github.com/anthropics/anthropic-sdk-go` | Anthropic Claude API client |

## Adding a New Tool

1. **Add method to `internal/tools/tools.go`** - Implement the tool logic
2. **Register in `cmd/screenshooter-mcp-server/main.go`** - Add input struct, call `mcp.AddTool()`
3. **Update tool list comments** - In main.go package doc and `registerTools()` doc
4. **Update `README.md`** - Add to features list and MCP Tools section
5. **Update `internal/tools/skills/SKILL.md`** - Add to tool catalog and workflows if applicable

### Pipeline Support

If the new tool should be usable inside `execute_capture_pipeline`, add a step executor in `internal/tools/pipeline.go`:
- Define `exec<ToolName>()` function
- Handle stack push/pop as appropriate (see stack behavior table in SKILL.md)
- Add the case to the `ExecutePipeline()` switch statement

## Packaging

Packages are built in `.github/workflows/packages.yml` for:

| Distribution | Package Format |
|--------------|----------------|
| Debian/Ubuntu | `.deb` |
| Fedora | `.rpm` |
| Alpine | `.tar.gz` (static) |

Each distribution has two variants:
- **server**: HTTP server with systemd unit, config in `/etc/screenshooter-mcp/`
- **stdio**: Standalone binary for MCP client integration

The SKILL.md file is installed to `/usr/share/screenshooter-mcp/skills/` in all packages.

## Git Workflow

- Create a feature branch from `main` for each independent change
- Commit with `git commit -s -S -m "subsystem: description"` (signoff + GPG sign)
- Use `subsystem: change description` title format (lowercase, concise)
- Commit message body should explain WHY, not HOW
