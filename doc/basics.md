# Project Structure

```
screenshooter-mcp/
├── cmd/screenshooter-mcp-server/  # Main entrypoint
├── internal/
│   ├── capture/              # Screen capture implementations
│   ├── config/               # Configuration
│   ├── logging/             # Logging
│   └── tools/               # MCP tools
└── .github/
    ├── workflows/           # CI/CD
    └── dependabot.yml       # Dependency updates
```

# Architecture

```
┌─────────────────────────────────────────────────────┐
│                    MCP Client                       │
│               (Claude Desktop, etc.)                │
└─────────────────────┬───────────────────────────────┘
                      │ stdio (MCP over stdin/stdout)
                      │ or HTTP (--listen flag)
┌─────────────────────▼────────────────────────────────┐
│                  MCP Server (Go)                     │
│  ┌─────────────┐  ┌─────────────┐  ┌───────────────┐ │
│  │   tools/    │  │   config/   │  │   capture/    │ │
│  │ capture_*   │──│             │──│   x11/        │ │
│  │ list_*      │  │             │  │   wayland/    │ │
│  └─────────────┘  └─────────────┘  └───────────────┘ │
└──────────────────────────────────────────────────────┘
```

# Dependencies

| Package | Purpose |
|---------|---------|
| `github.com/jessevdk/go-flags` | CLI argument parsing |
| `github.com/modelcontextprotocol/go-sdk` | MCP protocol |
| `github.com/rs/zerolog` | Structured logging |
| `github.com/nskaggs/perfuncted` | Screen capture (X11, Wayland, Portal) |
| `github.com/jezek/xgb` | X11 bindings for multi-monitor support |

