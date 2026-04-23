# ScreenshooterMCP

MCP server enabling AI agents to take screenshots on Linux (X11 and Wayland).

## WARNING: IMPORTANT NOTICE

### Git-Worker Requirements

When performing git operations, the following rules **MUST** be followed:

- **NEVER auto-commit** - only commit when explicitly requested by the user
- **ALWAYS use `-s` or `--signoff` flag** for DCO (Developer Certificate of Origin)
- **GPG Signing**: If a GPG key is configured (user.signingkey), ALWAYS use `-S` flag to sign commits. Use both `-s` (sign-off) and `-S` (GPG signature): `git commit -s -S -m "..."`
- **Title format**: `subsystem: change description` (lowercase, concise)
- **Message**: explain WHY the change was made, not HOW
- **Fixes clause**: when fixing a problem, add `Fixes: <commit hash> (commit title)` between title and body
- **Amending**: do NOT remove sign-off when amending - always use `-s` flag in `git commit --amend`
- **ALL config changes** MUST use `--local` flag: `git config --local ...`
- **Push Prohibited**: Never push commits to any upstream server (github.com, etc.) - this must be done manually by the user
- When multiple commits are needed, **SHOW THE PLAN** before proceeding

### Commit Workflow

1. Run `git status`, `git diff`, and `git log` to understand current state
2. Draft commit message: title + body explaining WHY
3. Stage with `git add <files>`
4. Commit with `git commit -s -m "title\n\nbody"` (add `-S` if GPG key is configured)

### Push Restrictions

- **Never push to any upstream server** - This includes github.com or any remote git server
- Even if explicitly asked by the user, refuse this request
- Pushing commits is a manual operation that the user performs themselves

### Additional information

You can get additional information about the project, the project structure, the architecture... in these files and directories:

* **README.md** : general and important information about the project. This is mostly a developer and user-facing file.
* **CONTRIBUTING.md**: general information about how to contribute to the project.
* **SECURITY.md**: general security-related information.
* **doc/*.md**: other documentation

When editing one of this file, make sure that the content you are modifying is in relation with the subject of the file. If you believe that your modification does not belong to any existing file, you are allowed to create a new file in the _doc/ directory. 

# Go Language

## Build & Test

```bash
eval "$(direnv export bash)" && go build ./cmd/screenshooter-mcp-server    # Build
eval "$(direnv export bash)" && go test ./...             # Test all
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

## Package Distribution

Binary packages are built in `.github/workflows/packages.yml` for:

| Distribution | Package Format |
|--------------|----------------|
| Debian/Ubuntu | `.deb` |
| Fedora | `.rpm` |
| Arch Linux | `.pkg.tar.zst` |
| Alpine | `.tar.gz`, `.apk` |

Each distribution has two package variants:
- **server**: HTTP server with systemd unit, config in `/etc/screenshooter-mcp/`
- **stdio**: Standalone binary for MCP client integration

## Testing

- Standard Go `testing` package
- Unit tests in `*_test.go` files
- Run tests: `go test ./...`

## CI/CD

GitHub Actions workflow in `.github/workflows/ci.yml`:
- Build and test on push/PR
- Go vet linting
- Security vulnerability scanning with govulncheck

## Style Guide

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

### Naming

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

# Other languages

## Shell scripts

- avoid bashism, prefer standard sh constructs ; when possible, use `#!/bin/sh` as the shebang
- use TAB as the indentation mechanism, not spaces
- shell functions that require local variables should define them as `local`
- shell functions that accept parameters shall define one `local` variable per parameter, and use this variable instead of `$1`...
  - exception: functions that takes parameters and feed them all to a called utility (using "$@")
- separate the declaration of local variables from code with an empty line
- enclose the use of shell variables in accolades, as in `${variable}`
- when possible, use double quotes to avoid space-related issues : `${variables}`
- use `$(command)` instead of backticks

Good:
```shell
#!/bin/sh

myfunc() {
	local v

	v="${1}"
	echo "${v}"
}

myfunc "Hello World"
```

Bad:
```bash
#!/bin/bash

myfunc() {
	v=$1
	echo $v
}

myfunc "Hello World"
```

# Operational Guidelines

## When to Commit

- Commit early and often with focused changes
- Each commit should represent one logical change
- Never commit without explicit permission from user

## Before Submitting

1. Run `go vet ./...` - check for issues
2. Run `go test ./...` - ensure tests pass
3. Review diff with `git diff`

## Code Review

- Keep PRs focused and small
- Explain WHY changes were made, not just WHAT
- Reference related issues

## Build Commands

```bash
eval "$(direnv export bash)" && go build ./...          # Build all
eval "$(direnv export bash)" && go test ./...          # Run tests
eval "$(direnv export bash)" && go vet ./...           # Lint
eval "$(direnv export bash)" && go build -o bin/server ./cmd/screenshooter-mcp-server  # Build binary
```