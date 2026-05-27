# Changelog

## ScreenshooterMCP v1.2.1

### What's Inside

#### Security Fixes

This release is a security release only. It fixes the following 

- [GO-2026-5026](https://pkg.go.dev/vuln/GO-2026-5026) (golang.org/x/net ; fixed in golang.org/x/net@0.55.0)
- [GO-2026-4982](https://pkg.go.dev/vuln/GO-2026-4982) (net ; fixed in net@1.26.3)
- [GO-2026-4980](https://pkg.go.dev/vuln/GO-2026-4980) (net ; fixed in net@1.26.3)
- [GO-2026-4971](https://pkg.go.dev/vuln/GO-2026-4971) (net ; fixed in net@1.26.3)
- [GO-2026-4918](https://pkg.go.dev/vuln/GO-2026-4918) (net ; fixed in net@1.26.3)

### Commits

#### v1.2.1

- cmd: bump version number to 1.2.1 (Emmanuel Deloget)
- doc: update CHANGELOG.md for version v1.2.1 (Emmanuel Deloget)
- go: update dependencies (Emmanuel Deloget)
- doc: update SECURITY.md to fix the version number (Emmanuel Deloget)
- build(deps): bump github.com/anthropics/anthropic-sdk-go (dependabot[bot])
- build(deps): bump github.com/nskaggs/perfuncted from 0.3.2 to 0.4.0 (dependabot[bot])
- build(deps): bump google.golang.org/genai from 1.56.0 to 1.57.0 (dependabot[bot])

#### v1.2.0

- cmd: bump version number to 1.2.0 (Emmanuel Deloget)

#### v1.2.0-rc0

- cmd: bump version number to v1.2.0-rc0
- go: update dependencies (Emmanuel Deloget)
- README: announce that tests on Arch Linux are successfull (Emmanuel Deloget)
- tests: add integration tests on Arch Linux (gnome+KDE, wayland only) (Emmanuel Deloget)
- packaging: make sure the user extension is installed using the right owner (Emmanuel Deloget)
- tests: update the integration README to re-order the various configurations (Emmanuel Deloget)
- tests: vm-lifecycle.sh learns command 'console' to start an interactive console in the VM (Emmanuel Deloget)
- tests: test-mcp shall authorize tools before it uses it (Emmanuel Deloget)
- README: announce support for Arch Linux packages (Emmanuel Deloget)
- workflow: add a build for Arch Linux (x86_64 and aarch64) (Emmanuel Deloget)
- tests: fix the building of the Arch package for arm64 (Emmanuel Deloget)
- tests: move the PKGBUILD.in file to its correct place (Emmanuel Deloget)
- tests: build a valid -arch package (Emmanuel Deloget)

## ScreenshooterMCP v1.1.0

### What's Inside

#### Vision Provider Fallback

When your primary AI vision provider is unavailable — a local Ollama instance that isn't running, or a cloud API when the network is down — ScreenshooterMCP now automatically falls back to the next configured provider. Enable it with `enable_fallback: true` in your config or `--enable-vision-fallback` on the command line. Providers are tried in configuration order until one succeeds, completely transparent to the caller.

#### Tool Access Control

Fine-grained control over which tools are available. Configure each tool with `"allow"`, `"deny"`, or `"ask"` policies in `config.json`. Tools with `"ask"` policy require explicit user authorization via `allow_tool_access`, which grants temporary access with a configurable expiration. 

Two new related tools have been added:

- `list_tool_access` — list the tool access permissions. 
- `allow_tool_access` — grant temporary permissions to a specific tool.

When it uses a tool which is not allowed by the user, the agents receives an error that tells it to ask for the permission to use this tool. This should help in environments where the agent tried to obtain access to a tool without asking the user first.

#### JSON Logging

Structured JSON log output is now available with `--log-format json` or `"log_format": "json"` in config. Perfect for log aggregation systems and automated log parsing.

#### New Vision Providers

A new native vision provider for Google Gemini API and Vertex AI. Supports both direct API key authentication and GCP Application Default Credentials for Vertex AI, giving you access to Gemini 2.5 Flash and other vision-capable models.

Information about how to set up NVIDIA NIM through the OpenAI-compatible provider has been added to `README.md`.

#### Code Refactoring

The codebase has been reorganized for better maintainability:
- **mcptools/** — Each MCP tool now lives in its own file under a dedicated package
- **access/** — Tool access control is managed by a dedicated `AccessManager`
- **utils/** — Shared utilities moved to their own package

### Commits

#### v1.1.0

- cmd: bump version number to v1.1.0
- go: update dependencies (Emmanuel Deloget)
- doc: add missing information about gemini in basics.md (Emmanuel Deloget)
- README: add some documentation for option --enable-vision-fallback (Emmanuel Deloget)
- vision: increase test coverage of vision.go (Emmanuel Deloget)
- vision: implement 'enable fallback' for image analysis and comparison (Emmanuel Deloget)
- cmd: the main binary learns option --enable-vision-fallback (Emmanuel Deloget)
- README: add doc about --log-format / log_format (Emmanuel Deloget)
- logging: set up JSON logging when --log-format JSON is used (Emmanuel Deloget)
- cmd: the MCP server learns a new --log-format option (Emmanuel Deloget)
- config: add a log_format configuration item (Emmanuel Deloget)
- cmd: remove the list of available tools (redundant information) (Emmanuel Deloget)
- access: do not construct the list of tools twice (Emmanuel Deloget)
- doc: fix all relevant doc to explain how access control is used (Emmanuel Deloget)
- opencode: add the so-called 'Andrej Karpathy' skill (Emmanuel Deloget)
- cmd: remove unused files (Emmanuel Deloget)
- cmd: use the tools in mcptools instead of defining them on-line (Emmanuel Deloget)
- config: add some configuration for tool access (Emmanuel Deloget)
- mcptools: implement all tools in a new package (Emmanuel Deloget)
- access: add a helper function to check if a tool has been registered (Emmanuel Deloget)
- utils: move all utils into a utils package (Emmanuel Deloget)
- access: implement an access control manager (Emmanuel Deloget)
- internal: fix some formating issues (Emmanuel Deloget)
- README: announce that we support Google Gemini as a vision provider (Emmanuel Deloget)
- vision: wire the Gemini vision provider (Emmanuel Deloget)
- vision: add some basic tests for the Gemini provider (Emmanuel Deloget)
- vision: add a new vision provider for Google Gemini (Emmanuel Deloget)
- go: update dependencies (Emmanuel Deloget)
- README: tell that we are compatible with NVIDIA NIM (Emmanuel Deloget)

## ScreenshooterMCP v1.0.0

### What's Inside

#### Screen Capture

ScreenshooterMCP automatically detects your display server (X11 or Wayland) and provides a unified set of capture tools:

- **list_monitors** — Discover all connected displays with their names, positions, and resolutions
- **capture_screen** — Grab the entire desktop or a specific monitor as a PNG
- **capture_window** — Capture any open window by title (partial matching supported), with window state detection (active, minimized, maximized, fullscreen)
- **capture_region** — Extract a precise rectangular area from the virtual screen

#### AI-Powered Vision Analysis

Configure your favorite AI vision providers and let agents understand what they see:

- **analyze_image** — Ask natural language questions about any screenshot
- **extract_text** — Pull structured Markdown text from images (OCR)
- **find_region** — Locate UI elements by description and get their bounding box coordinates
- **compare_images** — Spot differences between two screenshots
- **list_vision_providers** — See which AI providers are configured and ready

Three provider types are supported out of the box:

- **OpenAI-compatible** — Works with OpenAI, Ollama, Mistral, Groq, and any compatible API
- **Anthropic Claude** — Direct access to Claude's vision capabilities
- **HuggingFace** — Tap into the HuggingFace Inference API

#### Pipeline Execution

Chain multiple operations together in a single call with `execute_capture_pipeline`. Capture a window, find a specific element, zoom in on a region, and extract text — all in one request.

#### GNOME Shell Extension

For modern GNOME environments where direct window enumeration is restricted, we include a purpose-built GNOME Shell extension that exposes window management via D-Bus. It auto-detects your GNOME version and installs the compatible variant (legacy for GNOME 43-44, modern for GNOME 45+).

#### Flexible Deployment

- Stdio mode (default) — Drop-in integration with Claude Desktop, Cursor, and other MCP clients
- HTTP server mode — Run as a networked service with Streamable HTTP transport
- JSON configuration — Fine-tune logging, vision providers, and network settings via config files

#### Tested Across Distributions

ScreenshooterMCP has been validated through integration tests on Debian 12/13, Fedora 43, and Ubuntu 24.04/25.10 — across both GNOME and KDE, on both X11 and Wayland.

### Commits

#### v1.0.0

- cmd: bump version number to 1.0.0
- config: fix some config unit tests (Emmanuel Deloget)
- tools: don't allow the encoding of a nil image (Emmanuel Deloget)
- cmd: add more unit tests on the main program (Emmanuel Deloget)
- internal: improve test coverage (Emmanuel Deloget)
- go: update some go dependencies (Emmanuel Deloget)

#### v1.0.0-rc9

- cmd: bump version number to 1.0.0-rc9
- doc: add more developper oriented documentation (Emmanuel Deloget)
- cmd: the version is defined as a constant (Emmanuel Deloget)
- doc: version 0.1.0 never existed, 1.0.0 will (Emmanuel Deloget)
- README: update the test result section (Emmanuel Deloget)

#### v1.0.0-rc8

- README: fix the name qwen2-vl to qwen3-vl (Emmanuel Deloget)
- cmd: handle JSON double-serialization in execute_capture_pipeline (Emmanuel Deloget)
- tools: remove external file read from GetSkillInfo to prevent prompt injection (Emmanuel Deloget)
- docs: expand basics.md with comprehensive developer guide (Emmanuel Deloget)
- tools: add security obligations and workflow clarification to SKILL.md (Emmanuel Deloget)
- docs: add timeout parameter to vision tool docs in README (Emmanuel Deloget)
- docs: add get_skill_info_for_agent to README (Emmanuel Deloget)
- tools: add get_skill_info_for_agent tool with embedded SKILL.md (Emmanuel Deloget)
- packaging: install SKILL.md to /usr/share/screenshooter-mcp/skills/ (Emmanuel Deloget)
- tools: add SKILL.md agent usage guide (Emmanuel Deloget)
- docs: update tool lists in README and code comments (Emmanuel Deloget)
- tools: add execute_capture_pipeline for chaining operations (Emmanuel Deloget)
- go: update dependencies (Emmanuel Deloget)
- vision: add compare_images tool for two-image comparison (Emmanuel Deloget)
- tools: add timeout parameter to vision tool methods (Emmanuel Deloget)
- capture: add window state fields to Window struct (Emmanuel Deloget)
- docs: add model limitations warning to vision providers section (Emmanuel Deloget)
- gitignore: add config.json to ignored files (Emmanuel Deloget)
- tools: reformat vision prompts as multi-line strings (Emmanuel Deloget)
- tools: force find_region to return structured JSON (Emmanuel Deloget)
- vision: add info and debug logging to providers (Emmanuel Deloget)
- docs: update tool lists and add vision provider documentation (Emmanuel Deloget)
- main: integrate vision providers and MCP tools (Emmanuel Deloget)
- tools: add vision MCP tool methods (Emmanuel Deloget)
- vision: implement huggingface provider (Emmanuel Deloget)
- vision: implement anthropic provider (Emmanuel Deloget)
- vision: implement openai-compatible provider (Emmanuel Deloget)
- vision: add provider interface and manager (Emmanuel Deloget)
- config: add vision provider configuration structs (Emmanuel Deloget)
- README: use UTF-8 emoticons to represent PASS or FAIL (Emmanuel Deloget)
- workflow: fix the build of the static packages to properly add a README.md file (Emmanuel Deloget)

#### v1.0.0-rc7

- capture: update dependencies and implement the new IterateWindows method in the window.Manager interface (Emmanuel Deloget)
- docs: add integration test results table to README (Emmanuel Deloget)
- tests: fix the provisionning of KDE-based ubuntu VMs (Emmanuel Deloget)
- tests: simplify the provisionning of KDE-based debian VMs (Emmanuel Deloget)
- tests: provision-vm.sh functions may consume $version in addition to $disk and $mode (Emmanuel Deloget)
- tests/integration: remove ubuntu 25.10 kde x11 (Ubuntu 25.10 KDE is Wayland-only) (Emmanuel Deloget)
- tests/integration: fix Ubuntu KDE network by installing and enabling NetworkManager (Emmanuel Deloget)
- tests/integration: update configuration list for GNOME 49 and add Ubuntu 25.10 KDE (Emmanuel Deloget)
- tests/integration: fix Fedora KDE booting into GNOME instead of KDE (Emmanuel Deloget)
- tests/integration/shared/test-mcp: handle isError responses from tool calls (Emmanuel Deloget)
- tests/integration: fix all.sh loop consuming stdin (Emmanuel Deloget)
- tests: remove support for Ubuntu 26.04 (Emmanuel Deloget)
- tests: add a file that list the various test configurations (Emmanuel Deloget)
- tests: fix how KDE VM images are built (Emmanuel Deloget)
- tests: fix the slow boot on ubuntu VMs (Emmanuel Deloget)
- tests: some test steps are of interest to gnome only (Emmanuel Deloget)
- scripts: authorize-portal.sh shall work for gnome AND KDE (Emmanuel Deloget)
- README: add GNOME Shell Extension section after security notice (Emmanuel Deloget)

#### v1.0.0-rc6

- workflow: the -stdio packages may also require the use of the gnome extension (Emmanuel Deloget)
- scripts: add packaging files for the -stdio packages on debian and fedora (Emmanuel Deloget)
- workflow, tests: fix a typo on the copy of prerm files (Emmanuel Deloget)
- workflow: add the gnome extension in the static builds as well (Emmanuel Deloget)
- workflow, tests: use the existing scripts and files instead of heredoc for fedora packages (Emmanuel Deloget)
- scripts: add fedora-related postinst/prerm scripts (Emmanuel Deloget)
- workflow, tests: use the existing scripts and files instead of heredoc for debian packages (Emmanuel Deloget)
- scripts: prepare the installable scripts and files (Emmanuel Deloget)
- capture: try to use our own window manager when we cannot get one from perfuncted (Emmanuel Deloget)
- capture: add a new window manager to use our own gnome extension if needed (Emmanuel Deloget)
- test: unlog/relog the user during integration tests (Emmanuel Deloget)
- tests: fix the test creation of both debian/ubuntu and fedora packages to install the gnome extension (Emmanuel Deloget)
- gnome: add a small and rough gnome extension to list and manipulate windows (Emmanuel Deloget)
- tests: run.sh learns option --no-test (Emmanuel Deloget)
- tests: vm-lifecycle.sh lears a new command 'view' (Emmanuel Deloget)
- tests: remove unneeded commented code (Emmanuel Deloget)
- build(deps): Bump github.com/nskaggs/perfuncted from 0.1.8 to 0.1.9 (dependabot[bot])
- build(deps): Bump softprops/action-gh-release from 2 to 3 (dependabot[bot])
- build(deps): Bump actions/download-artifact from 4 to 8 (dependabot[bot])

#### v1.0.0-rc5

- ci: update package builds to use ExecStartPre for portal authorization (Emmanuel Deloget)
- tests: add local packaging scripts for deb and rpm builds (Emmanuel Deloget)
- go: update dependencies (Emmanuel Deloget)

#### v1.0.0-rc4

- README: warn about automatic portal authorization in server packages (Emmanuel Deloget)
- ci: add portal authorization script and service to packages (Emmanuel Deloget)

#### v1.0.0-rc3

- ci: remove dconf configuration from VM provisioning (Emmanuel Deloget)
- tests: run all integration test steps even on test-mcp failure (Emmanuel Deloget)

#### v1.0.0-rc2

- capture: allow degraded mode when window backend unavailable on Wayland (Emmanuel Deloget)
- README: add testing section linking to integration tests (Emmanuel Deloget)
- tests: add integration tests README (Emmanuel Deloget)
- tests: add Ubuntu Desktop support with cloud-init autoinstall (Emmanuel Deloget)
- tests: use GitHub tags API instead of releases API (Emmanuel Deloget)
- tests: fix Fedora image creation and provisioning (Emmanuel Deloget)
- tests: reject Fedora 43 X11 combinations (not supported) (Emmanuel Deloget)
- tests: update download-iso.sh with latest stable releases (Emmanuel Deloget)
- tests: drop Fedora 42 support, keep only Fedora 43 (Emmanuel Deloget)
- tests: add KDE Wayland support to provisioning scripts (Emmanuel Deloget)
- tests: fix SSE parser for multi-line events and large images (Emmanuel Deloget)
- tests: fix image data parsing to handle base64 encoding (Emmanuel Deloget)
- tests: get session ID from server response header (Emmanuel Deloget)
- tests: add MCP session header to test-mcp requests (Emmanuel Deloget)
- tests: rewrite test-mcp with proper MCP protocol handling (Emmanuel Deloget)
- tests: add MCP initialize handshake to test-mcp (Emmanuel Deloget)
- tests: handle SSE response in test-mcp (Emmanuel Deloget)
- tests: add shared test-mcp tool to repository (Emmanuel Deloget)
- tests: adapt integration tests for user service (Emmanuel Deloget)
- tests: vm-lifeycle.sh shall operate on a system vm, not a session vm (Emmanuel Deloget)
- tests: implement some integration test on all supported platforms (Emmanuel Deloget)

#### v1.0.0-rc1

- ci: switch server package to systemd user service (Emmanuel Deloget)
- capture: add loginctl fallback for environment detection (Emmanuel Deloget)
- docs: add some shell programming style in AGENTS.md (Emmanuel Deloget)
- go.mod: update dependencies (Emmanuel Deloget)
- x11: activate window before capture to get geometry (Emmanuel Deloget)
- docs: refresh README with current package availability (Emmanuel Deloget)

#### v1.0.0-rc0

- ci: remove packaging for Arch Linux (Emmanuel Deloget)
- ci: fix Arch X86_64 to x86_64 (Emmanuel Deloget)
- ci: remove the build of aarch64 for now (Emmanuel Deloget)
- ci: remove the Alpine APK build for now (Emmanuel Deloget)
- ci: add some debug on Arch Linux build (Emmanuel Deloget)
- ci: rename the alpine build as a static build (Emmanuel Deloget)
- ci: better tentative to fix Arch Linux builds (Emmanuel Deloget)
- ci: tentative fix for Arch Linus builds (Emmanuel Deloget)
- ci: try to fix the fedora/rpm build and to provide correct postinst/prerm (Emmanuel Deloget)
- ci: debian control files are NOT created in pkg/ (Emmanuel Deloget)
- ci: try to fix the debian build and to provide correct postinst/prerm (Emmanuel Deloget)
- ci: make sure both -buildvcs and -trimpath options are used when doing a go build (Emmanuel Deloget)
- ci: Step 2 - Fix Fedora, Alpine and Arch builds (Emmanuel Deloget)
- ci: Step 1 - Universal fixes for all platforms (Emmanuel Deloget)
- ci: fix Arch and Alpine package build issues (Emmanuel Deloget)
- ci: various packages workflow improvements (Emmanuel Deloget)
- ci: fix systemd service stop in before-remove scripts (Emmanuel Deloget)
- ci: add -buildvcs=false to go build commands (Emmanuel Deloget)
- ci: fix packages workflow build configurations (Emmanuel Deloget)
- ci: install govulncheck before running security scan (Emmanuel Deloget)
- ci: add release job to attach artifacts to GitHub release (Emmanuel Deloget)
- deps: update perfuncted and fix context requirements (Emmanuel Deloget)
- ci: auto-trigger packages workflow on GitHub release (Emmanuel Deloget)
- docs: add GPG signing and push restriction rules (Emmanuel Deloget)
- deps: add govulncheck for security scanning in CI (Emmanuel Deloget)
- docs: add godoc comments to all exported functions and types (Emmanuel Deloget)
- docs: add Ideas for Contributions section (Emmanuel Deloget)
- docs: update AGENTS.md and CONTRIBUTING.md (Emmanuel Deloget)
- docs: reorganize documentation files (Emmanuel Deloget)
- docs: add CI badges to README (Emmanuel Deloget)
- docs: add issue and PR templates (Emmanuel Deloget)
- docs: add CONTRIBUTING.md with contribution guidelines (Emmanuel Deloget)
- docs: add SECURITY.md with vulnerability reporting policy (Emmanuel Deloget)
- ci: add Dependabot for dependency updates (Emmanuel Deloget)
- docs: add MIT license (Emmanuel Deloget)
- docs: add README with installation and usage guide (Emmanuel Deloget)
- config: add Listen field and fix XDG config loading (Emmanuel Deloget)
- docs: update AGENTS.md to reflect current architecture (Emmanuel Deloget)
- capture: update X11/Wayland capture implementation (Emmanuel Deloget)
- test: add integration test script (Emmanuel Deloget)
- tools: refactor capture methods to return []byte (Emmanuel Deloget)
- refactor: remove vision bundling, focus on core screenshot tools (Emmanuel Deloget)
- gitignore: add curl.* and server.log temporary files (Emmanuel Deloget)
- config: add JSON configuration system (Emmanuel Deloget)
- server: basic implementation for all tools (Emmanuel Deloget)
- build: add cleanup to build script (Emmanuel Deloget)
- build: fix OLLAMA_LIBRARY_PATH for Vulkan GPU support (Emmanuel Deloget)
- server: make vision model default based on quality (Emmanuel Deloget)
- build: remove MLX CUDA library from AppImage (Emmanuel Deloget)
- build: improve AppImage and Ollama integration (Emmanuel Deloget)
- build: fix Python 3.6 compatibility in icon converter (Emmanuel Deloget)
- gitignore: add *.AppImage to ignore AppImage files (Emmanuel Deloget)
- vision: fix Ollama path detection for AppImage (Emmanuel Deloget)
- build: add AppImage build infrastructure (Emmanuel Deloget)
- refactor: rename cmd/server to cmd/screenshooter-mcp-server (Emmanuel Deloget)
- docs: add amending rule for git commits (Emmanuel Deloget)
- server: add --listen flag with Streamable HTTP transport (Emmanuel Deloget)
- vision: improve Ollama startup and environment configuration (Emmanuel Deloget)
- logging: add logging throughout the codebase (Emmanuel Deloget)
- logging: add zerolog-based logging package (Emmanuel Deloget)
- go: add dependencies for logging (Emmanuel Deloget)
- docs: update AGENTS.md with project implementation details (Emmanuel Deloget)
- server: add MCP server entry point with CLI flags (Emmanuel Deloget)
- tools: add MCP tools for screenshot and element capture (Emmanuel Deloget)
- vision: add Ollama manager for vision model integration (Emmanuel Deloget)
- capture: add Wayland screen capture implementation (Emmanuel Deloget)
- capture: add X11 screen capture implementation (Emmanuel Deloget)
- capture: add core types and environment detection (Emmanuel Deloget)
- ci: add GitHub Actions workflow for CI (Emmanuel Deloget)
- go: initialize module github.com/emmanuel-deloget/screenshooter-mcp (Emmanuel Deloget)
- git: add bin/ to .gitignore (Emmanuel Deloget)
- go: add go.mod and .envrc for local Go environment (Emmanuel Deloget)
- docs: add Go development environment details to AGENTS.md (Emmanuel Deloget)
- git: add .gitignore for Go project (Emmanuel Deloget)
- agents: add optimization-expert agent (Emmanuel Deloget)
- agents: add image-reader agent (Emmanuel Deloget)
- agents: add architecture-expert agent (Emmanuel Deloget)
- skills: add uml-worker skill (Emmanuel Deloget)
- skills: add todo-manager skill (Emmanuel Deloget)
- skills: add image-worker skill (Emmanuel Deloget)
- skills: add git-worker skill (Emmanuel Deloget)
- skills: add brainstorming skill (Emmanuel Deloget)
- docs: add AGENTS.md with project documentation (Emmanuel Deloget)
