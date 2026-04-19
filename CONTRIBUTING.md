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

## Questions

For questions, open a discussion on GitHub.

## Ideas for Contributions

The project is focused on screenshot capabilities for Linux. Contributions are welcome in these areas:

### Platform Support

- Support for other operating systems (macOS, Windows)
- Support for additional Linux distributions

### Architecture Support

- Support for additional CPU architectures (ARM, RISC-V)

### Related Tools

- Screen annotation or drawing tools
- Image cropping or manipulation
- Any tools that enhance the screenshot workflow

### Vision Model Integration

- Integration with vision models (local or API-based) for element detection
- OCR capabilities for captured screens

**Note:** Currently, only X11 and Wayland on Linux are supported. If you'd like to add support for other platforms, please open an issue to discuss the approach first.