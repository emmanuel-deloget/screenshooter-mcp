# Development Process

This document describes the development workflow, coding standards, and operational procedures for ScreenshooterMCP.

## Build Commands

```bash
# Load direnv environment
eval "$(direnv export bash)"

# Build the server
go build ./cmd/screenshooter-mcp-server

# Build all packages
go build ./...

# Build binary to specific location
go build -o bin/server ./cmd/screenshooter-mcp-server

# Run all tests
go test ./...

# Run specific package tests
go test ./internal/tools/...

# Run with coverage
go test -cover ./...

# Lint
go vet ./...

# Format code
go fmt ./...
```

## Development Environment

- **Module**: `github.com/emmanuel-deloget/screenshooter-mcp`
- **Go Version**: 1.26 or later
- **Vendoring**: Not used
- **Local GOPATH**: Managed via `.envrc` with direnv
  - Modules cached in `./.go/pkg/mod`
  - Binaries installed to `./.go/bin`

## Git Workflow

### Branch Strategy

- Create feature branches from `main` for each independent change
- Use descriptive branch names: `feature/<name>`, `fix/<name>`, `docs/<name>`

### Commit Requirements

**MUST follow these rules:**

- **ALWAYS use `-s` or `--signoff` flag** for DCO (Developer Certificate of Origin)
- **GPG Signing**: If a GPG key is configured (`user.signingkey`), ALWAYS use `-S` flag to sign commits
  - Use both `-s` (sign-off) and `-S` (GPG signature): `git commit -s -S -m "..."`
- **Title format**: `subsystem: change description` (lowercase, concise)
- **Message**: explain WHY the change was made, not HOW
- **Amending**: do NOT remove sign-off when amending - always use `-s` (and maybe `-S`) flags in `git commit --amend`

## Go Code Style Guide

### General Principles

- Keep functions focused and small
- Use meaningful variable names
- Avoid global state
- Return errors explicitly, don't use panic
- Prefer clear over clever

### Error Handling

```go
// Good: explicit error handling
if err != nil {
    return fmt.Errorf("failed to create capture: %w", err)
}

// Bad: ignoring errors
_ = something()
```

### Naming Conventions

- Use camelCase for variables and functions
- Use PascalCase for exported types and functions
- Use snake_case for file names
- Keep names descriptive but not verbose

### Imports

- Group stdlib imports separately from external packages
- Use meaningful aliases only when needed

```go
import (
    "context"
    "fmt"

    "github.com/example/package"
)
```

### Formatting

- Always run `gofmt` or `go fmt ./...` before committing
- Use `gofmt -w .` to format automatically
- Don't fight gofmt - follow its conventions

### Long Lines

When splitting function arguments across lines, each argument goes on its own line:

```go
// Good
value := SomeFunctionCall(
    theFirstArgument,
    theSecondArgument,
    theThirdArgument,
    somePrivateFunctionCall(),
    anotherArgument,
)

// Bad - arguments on same line
value := SomeFunctionCall(
    arg1, arg2, arg3,
)
```

## Shell Script Guidelines

- Avoid bashism, prefer standard sh constructs
- Use `#!/bin/sh` as the shebang when possible
- Use TAB as the indentation mechanism, not spaces
- Shell functions that require local variables should define them as `local`
- Shell functions that accept parameters shall define one `local` variable per parameter
- Separate the declaration of local variables from code with an empty line
- Enclose shell variables in accolades: `${variable}`
- Use double quotes to avoid space-related issues
- Use `$(command)` instead of backticks

## Testing

### Test Structure

- Place tests in `*_test.go` files in the same package
- Use table-driven tests for multiple test cases
- Name test functions with `Test` prefix
- Test behavior, not implementation

```go
func TestCaptureScreen(t *testing.T) {
    tests := []struct {
        name    string
        monitor string
        want    error
    }{...}

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // test logic
        })
    }
}
```

### Test Coverage Goals

- Aim for >80% coverage on core packages
- Mock external dependencies (capture backends, vision APIs)
- See `doc/unit-testing-approach.md` for detailed testing strategies

## Before Submitting

1. Run `go vet ./...` - check for issues
2. Run `go test ./...` - ensure tests pass
3. Run `go fmt ./...` - ensure formatting is correct
4. Review diff with `git diff`

## CI/CD

GitHub Actions workflow in `.github/workflows/ci.yml`:
- Build and test on push/PR
- Go vet linting
- Security vulnerability scanning with govulncheck

## Package Distribution

Binary packages are built in `.github/workflows/packages.yml` for:

| Distribution | Package Format |
|--------------|----------------|
| Debian/Ubuntu | `.deb` |
| Fedora | `.rpm` |

Each distribution has two package variants:
- **server**: HTTP server with systemd unit, config in `/etc/screenshooter-mcp/`
- **stdio**: Standalone binary for MCP client integration

## Release Process

1. Update version strings in `cmd/screenshooter-mcp-server/main.go`
2. Update `CHANGELOG.md` with release notes
3. Create and push a tag: `git tag vX.Y.Z && git push origin vX.Y.Z`
4. GitHub Actions will build and publish packages
5. Create GitHub Release with release notes from CHANGELOG

## Code Review Guidelines

- Keep PRs focused and small
- Explain WHY changes were made, not just WHAT
- Reference related issues
- Ensure tests pass before requesting review
