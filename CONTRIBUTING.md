# Contributing to ScreenshooterMCP

Thank you for your interest in contributing!

## Ways to Contribute

- Report bugs
- Suggest new features
- Improve documentation
- Submit pull requests

## Reporting Bugs

Before reporting a bug:

1. Search existing issues to see if it's already reported
2. Use the bug report template if available

**For security vulnerabilities**, follow the procedure in [SECURITY.md](SECURITY.md).

Include:
- Go version (`go version`)
- Your OS and desktop environment (X11/Wayland)
- Steps to reproduce
- Actual vs expected behavior

## AI-Assisted Contributions

AI-assisted contributions (e.g., using AI agents to help write code) are welcome, provided:

- A human user verifies all proposed changes before submitting
- You understand and can explain the code being submitted
- You take responsibility for the quality of the code

**PRs automatically submitted by AI agents without human verification will be closed immediately.**

## Submitting Pull Requests

### Prerequisites

- Go 1.26 or later
- Git

### Development Setup

```bash
git clone https://github.com/emmanuel-deloget/screenshooter-mcp.git
cd screenshooter-mcp
go build ./...
go test ./...
```

### Coding Standards

- Run `go vet ./...` before committing
- Add tests for new features
- Keep changes focused and atomic

### Commit Messages

Format: `subsystem: change description`

Examples:
```
config: add Listen field
ci: add multi-distro package build workflow
docs: update README with installation instructions
```

### Submitting

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/my-feature`
3. Make changes and commit
4. Push to your fork
5. Open a pull request

## Project Structure

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
    └── dependabot.yml        # Dependency updates
```

## Questions

For questions, open a discussion on GitHub.