// Copyright 2025 Emmanuel Deloget. All rights reserved.
// Use of this source code is governed by the license that can be found in the LICENSE file.

// Package main provides the MCP server implementation for capturing screenshots on Linux.
//
// This server implements the Model Context Protocol (MCP) and exposes tools for
// capturing screens, windows, and regions on Linux systems running either X11 or
// Wayland desktop environments. It can operate in two modes:
//
//   - stdio mode: Communicates with an MCP client via standard input/output
//   - HTTP mode: Exposes an HTTP endpoint for MCP client connections
//
// Configuration is loaded from JSON files, following XDG Base Directory specification.
// The server will look for configuration in the following locations (in order of precedence):
//
//  1. Path specified via --config command-line flag
//  2. Path in SCREENSHOOTER_CONFIG environment variable
//  3. $XDG_CONFIG_HOME/screenshooter-mcp/config.json (or ~/.config/screenshooter-mcp/config.json)
//  4. /etc/screenshooter-mcp/config.json
//
// Example config.json:
//
//	{
//	  "log_level": "info",
//	  "color": "auto",
//	  "listen": "127.0.0.1:11777"
//	}
//
// Available tools: the list of available tools is defined in registerTools().
//
// Usage:
//
//	# Run in stdio mode (default)
//	screenshooter-mcp-server
//
//	# Run as HTTP server
//	screenshooter-mcp-server --listen 127.0.0.1:11777
//	screenshooter-mcp-server --stdio
//
//	# With custom config
//	screenshooter-mcp-server --config /path/to/config.json
//
//	# With logging
//	screenshooter-mcp-server --log-level debug
package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/emmanuel-deloget/screenshooter-mcp/cmd/screenshooter-mcp-server/access"
	"github.com/emmanuel-deloget/screenshooter-mcp/cmd/screenshooter-mcp-server/mcptools"
	"github.com/emmanuel-deloget/screenshooter-mcp/cmd/screenshooter-mcp-server/utils"
	"github.com/emmanuel-deloget/screenshooter-mcp/internal/capture"
	"github.com/emmanuel-deloget/screenshooter-mcp/internal/capture/wayland"
	"github.com/emmanuel-deloget/screenshooter-mcp/internal/capture/x11"
	"github.com/emmanuel-deloget/screenshooter-mcp/internal/config"
	"github.com/emmanuel-deloget/screenshooter-mcp/internal/logging"
	"github.com/emmanuel-deloget/screenshooter-mcp/internal/tools"
	"github.com/emmanuel-deloget/screenshooter-mcp/internal/vision"
	"github.com/jessevdk/go-flags"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Options defines the command-line flags and configuration overrides accepted by the server.
//
// The Options struct uses the go-flags library to parse command-line arguments.
// Each field corresponds to a command-line flag that can be passed when starting
// the server. Fields marked with jsonschema tags are used for generating
// JSON Schema documentation for the MCP tools.
//
// The --config flag allows specifying a custom path to a configuration file.
// If not provided, configuration is loaded from standard XDG locations.
//
// The --log-level flag controls the verbosity of logging output. Valid values are:
//   - debug: Most verbose, includes detailed debug information
//   - info: Default, includes operational information
//   - warn: Only warnings and errors
//   - error: Only errors
//
// The --color flag controls whether the logger uses colored output. Valid values are:
//   - always: Always use ANSI color codes
//   - never: Never use color codes
//   - auto: Detect if terminal supports colors (default)
//
// The --listen flag specifies the TCP address to listen on for HTTP mode.
// Use "stdio" as the value to communicate via standard input/output instead.
// The HTTP mode requires an external MCP<->HTTP bridge to convert between
// HTTP and the MCP stdio protocol.
//
// The --stdio flag is a convenience flag that forces stdio mode, equivalent
// to setting --listen to "stdio". It overrides any --listen value.
type Options struct {
	Version              bool   `short:"v" long:"version" description:"Show version"`
	Help                 bool   `short:"h" long:"help" description:"Show help"`
	Config               string `long:"config" description:"Path to config file"`
	LogLevel             string `short:"l" long:"log-level" description:"Log level" default:"info"`
	LogFormat            string `long:"log-format" description:"Log format: text|json" default:"text"`
	Color                string `long:"color" description:"Color output: always|never|auto" default:"auto"`
	Listen               string `long:"listen" description:"Listen on TCP address (e.g. 127.0.0.1:11777) or 'stdio' for stdio mode" default:""`
	Stdio                bool   `long:"stdio" description:"Run in stdio mode (overrides --listen)"`
	EnableVisionFallback bool   `long:"enable-vision-fallback" description:"Enable automatic fallback to next vision provider on error or timeout"`
}

const (
	ScreenshooterMCPVersion = "v1.2.0"
)

func main() {
	opts := Options{}
	parser := flags.NewParser(&opts, flags.Default)
	parser.Usage = "[options]"

	if _, err := parser.Parse(); err != nil {
		if flagsErr, ok := err.(*flags.Error); ok {
			if flagsErr.Type == flags.ErrHelp {
				os.Exit(0)
			}
			if flagsErr.Type == flags.ErrCommandRequired {
				_, _ = fmt.Fprintln(os.Stderr, "Error: command required")
				os.Exit(1)
			}
		}
		os.Exit(1)
	}

	cfg, err := config.Load(opts.Config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	if opts.LogLevel != "info" {
		cfg.LogLevel = opts.LogLevel
	}
	if opts.LogFormat != "text" {
		cfg.LogFormat = opts.LogFormat
	}
	if opts.Color != "auto" {
		cfg.Color = opts.Color
	}
	if opts.Listen != "" {
		cfg.Listen = opts.Listen
	}
	if opts.EnableVisionFallback {
		if cfg.Vision != nil {
			cfg.Vision.EnableFallback = true
		}
	}

	logging.Init(cfg.LogLevel, cfg.Color, cfg.LogFormat)

	if opts.Help {
		parser.WriteHelp(os.Stdout)
		os.Exit(0)
	}

	if opts.Version {
		fmt.Println("screenshot-mcp-server version " + ScreenshooterMCPVersion)
		os.Exit(0)
	}

	if err := run(&opts, cfg); err != nil {
		logging.Error().Err(err).Msg("Server error")
		os.Exit(1)
	}
}

// run starts the MCP server in stdio mode.
//
// This function is the main entry point for running the server. It detects the current
// desktop environment (X11 or Wayland), creates an appropriate screen capture backend,
// registers the MCP tools, and starts the server running on stdio transport.
//
// The detection process checks the XDG_SESSION_TYPE environment variable first, then falls
// back to checking for DISPLAY (X11) or WAYLAND_DISPLAY (Wayland) environment variables.
// If neither desktop environment is detected, an error is returned.
//
// The function creates platform-specific capture implementations:
//   - For X11: Uses xgb for RANDR monitor enumeration and perfuncted for capture
//   - For Wayland: Uses perfuncted (portal-based) for capture
//
// Once the capture backend is created, all MCP tools are registered:
// list_monitors, list_windows, capture_screen, capture_window, capture_region,
// and vision tools (list_vision_providers, analyze_image, extract_text, find_region)
// if vision providers are configured.
// The server then runs indefinitely, processing MCP requests via stdio.
//
// Returns an error if:
//   - The desktop environment cannot be detected
//   - The capture backend cannot be created
//   - The server fails to run
func run(opts *Options, cfg *config.Config) error {
	// Use config listen address, or fallback to stdio
	listen := cfg.Listen
	if opts.Stdio {
		listen = "stdio"
	} else if opts.Listen != "" {
		listen = opts.Listen
	}

	if listen != "" && listen != "stdio" {
		logging.Warn().Str("listen", listen).Msg("Listen mode: requires external MCP<->HTTP bridge")
		opts.Listen = listen
		return runHttpBridge(opts, cfg)
	}

	logging.Info().Msg("Starting screenshooter-mcp server")

	detector := capture.NewEnvironmentDetector()
	env, err := detector.Detect()
	if err != nil {
		logging.Error().Err(err).Msg("Failed to detect environment")
		return fmt.Errorf("failed to detect environment: %w", err)
	}
	logging.Info().Str("environment", string(env)).Msg("Environment detected")

	var capt capture.ScreenCapture
	switch env {
	case capture.EnvironmentX11:
		logging.Debug().Msg("Creating X11 capture")
		capt, err = x11.NewX11Capture()
		if err != nil {
			logging.Error().Err(err).Msg("Failed to create X11 capture")
			return fmt.Errorf("failed to create X11 capture: %w", err)
		}
	case capture.EnvironmentWayland:
		logging.Debug().Msg("Creating Wayland capture")
		capt, err = wayland.NewWaylandCapture()
		if err != nil {
			logging.Error().Err(err).Msg("Failed to create Wayland capture")
			return fmt.Errorf("failed to create Wayland capture: %w", err)
		}
	default:
		return fmt.Errorf("unsupported environment: %s", env)
	}

	serverTools := tools.NewTools(capt)

	visionMgr, err := vision.NewManager(cfg.Vision)
	if err != nil {
		logging.Warn().Err(err).Msg("Failed to initialize vision providers")
	} else if visionMgr != nil {
		serverTools.SetVisionManager(visionMgr)
		logging.Info().Int("count", len(cfg.Vision.Providers)).Msg("Vision providers initialized")
	} else {
		logging.Info().Msg("No vision providers configured")
	}

	accessMgr := access.NewAccessManager(cfg.Access, cfg.TempAccessDuration)

	server := mcp.NewServer(&mcp.Implementation{
		Name:    "screenshooter-mcp",
		Version: ScreenshooterMCPVersion,
	}, nil)

	registerTools(server, serverTools, accessMgr)

	logging.Info().Msg("MCP server running on stdio")
	return server.Run(context.Background(), &mcp.StdioTransport{})
}

// runHttpBridge starts the MCP server in HTTP mode.
//
// This function runs the server as an HTTP server, using the MCP SDK's
// StreamableHTTPHandler to handle client connections. The server listens
// on the TCP address specified in opts.Listen.
//
// HTTP mode is useful when the MCP client cannot communicate via stdio,
// such as when running the server as a remote service. However, MCP
// clients typically expect stdio communication, so HTTP mode requires
// an external MCP<->HTTP bridge to translate between HTTP and the MCP protocol.
//
// The detection of the desktop environment and creation of the capture backend
// follows the same process as the stdio mode (see run function). Once the
// server is configured, it starts listening on the specified address
// and handles incoming HTTP requests.
//
// Common use cases:
//   - Running behind a reverse proxy
//   - Containerized deployments
//   - Remote MCP server access
//
// Returns an error if:
//   - The desktop environment cannot be detected
//   - The capture backend cannot be created
//   - The HTTP server fails to start or listen
func runHttpBridge(opts *Options, cfg *config.Config) error {
	logging.Info().Str("listen", opts.Listen).Msg("Starting HTTP server")

	detector := capture.NewEnvironmentDetector()
	env, err := detector.Detect()
	if err != nil {
		return fmt.Errorf("failed to detect environment: %w", err)
	}

	var capt capture.ScreenCapture
	switch env {
	case capture.EnvironmentX11:
		logging.Debug().Msg("Creating X11 capture")
		capt, err = x11.NewX11Capture()
		if err != nil {
			logging.Error().Err(err).Msg("Failed to create X11 capture")
			return fmt.Errorf("failed to create X11 capture: %w", err)
		}
	case capture.EnvironmentWayland:
		logging.Debug().Msg("Creating Wayland capture")
		capt, err = wayland.NewWaylandCapture()
		if err != nil {
			logging.Error().Err(err).Msg("Failed to create Wayland capture")
			return fmt.Errorf("failed to create Wayland capture: %w", err)
		}
	default:
		return fmt.Errorf("unsupported environment: %s", env)
	}

	server := mcp.NewServer(&mcp.Implementation{
		Name:    "screenshooter-mcp",
		Version: ScreenshooterMCPVersion,
	}, nil)

	serverTools := tools.NewTools(capt)

	visionMgr, err := vision.NewManager(cfg.Vision)
	if err != nil {
		logging.Warn().Err(err).Msg("Failed to initialize vision providers")
	} else if visionMgr != nil {
		serverTools.SetVisionManager(visionMgr)
	}

	accessMgr := access.NewAccessManager(cfg.Access, cfg.TempAccessDuration)

	registerTools(server, serverTools, accessMgr)

	handler := mcp.NewStreamableHTTPHandler(func(req *http.Request) *mcp.Server {
		return server
	}, nil)

	logging.Info().Str("listen", opts.Listen).Msg("HTTP server listening")
	return http.ListenAndServe(opts.Listen, handler)
}

// registerTools registers all MCP tools with the server.
//
// Each tool is wrapped with error handling that:
//   - Logs the tool call with parameters for debugging
//   - Converts errors to user-friendly error messages
//   - Returns appropriate MCP content (text for errors, image for success)
//
// The tools use the ScreenCapture interface from the capture package, which
// provides a unified API regardless of the underlying desktop environment
// (X11 or Wayland). This abstraction allows the MCP tools to work
// identically regardless of which backend is in use.
//
// Tool result format:
//   - On success: Returns image data as ImageContent (PNG format) or JSON text
//   - On error: Returns error message as TextContent with IsError flag set
//
// The function logs at info level the names of all registered tools for
// verification purposes.
func registerTools(server *mcp.Server, t *tools.Tools, am *access.AccessManager) {
	checkAccess := func(tool string) *mcp.CallToolResult {
		return utils.CheckAccess(tool, am)
	}

	// --- Exempt tools (always allowed) ---

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_monitors",
		Description: "List all available monitors with their names and aliases",
	}, mcptools.ListMonitors(t))
	am.RegisterTool("list_monitors", true)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_vision_providers",
		Description: "List all configured AI vision providers",
	}, mcptools.ListVisionProvides(t))
	am.RegisterTool("list_vision_providers", true)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_skill_info_for_agent",
		Description: "Return the agent skill documentation for using this MCP server. Provides tool descriptions, workflow examples, and pipeline usage guidance.",
	}, mcptools.GetSkillInfoForAgent(t))
	am.RegisterTool("get_skill_info_for_agent", true)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_tool_access",
		Description: "List all tools with their current access status (allow, deny, ask).",
	}, mcptools.ListToolAccess(am))
	am.RegisterTool("list_tool_access", true)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "allow_tool_access",
		Description: "Grant temporary access to a tool that has 'ask' policy. Access is granted for a limited time.",
	}, mcptools.AllowToolAccess(am))
	am.RegisterTool("allow_tool_access", true)

	// --- Access-controlled tools ---

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_windows",
		Description: "List all open windows with their titles and IDs",
	}, mcptools.ListWindows(checkAccess, t))
	am.RegisterTool("list_windows", false)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "capture_screen",
		Description: "Capture the full screen or a specific monitor",
	}, mcptools.CaptureScreen(checkAccess, t))
	am.RegisterTool("capture_screen", false)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "capture_window",
		Description: "Capture a specific window by its title (partial match supported)",
	}, mcptools.CaptureWindow(checkAccess, t))
	am.RegisterTool("capture_window", false)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "capture_region",
		Description: "Capture a region from the virtual screen",
	}, mcptools.CaptureRegion(checkAccess, t))
	am.RegisterTool("capture_region", false)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "analyze_image",
		Description: "Analyze an image using AI vision providers",
	}, mcptools.AnalyzeImage(checkAccess, t))
	am.RegisterTool("analyze_image", false)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "extract_text",
		Description: "Extract text from an image as formatted markdown",
	}, mcptools.ExtractText(checkAccess, t))
	am.RegisterTool("extract_text", false)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "find_region",
		Description: "Find bounding box coordinates of a described element in an image",
	}, mcptools.FindRegion(checkAccess, t))
	am.RegisterTool("find_region", false)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "compare_images",
		Description: "Compare two images and describe the differences",
	}, mcptools.CompareImages(checkAccess, t))
	am.RegisterTool("compare_images", false)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "execute_capture_pipeline",
		Description: "Execute a pipeline of capture and vision operations. Each step's output is pushed onto a stack for use by subsequent steps. Returns the final result.",
	}, mcptools.ExecuteCapturePipeline(checkAccess, t))
	am.RegisterTool("execute_capture_pipeline", false)

	logging.Info().Strs("tools", am.ListTools()).Msg("Tools registered")
}
