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
// Available tools:
//
//   - list_monitors: Lists all available monitors with their names and aliases
//   - list_windows: Lists all open windows with their titles and IDs
//   - capture_screen: Captures the full screen or a specific monitor
//   - capture_window: Captures a specific window by its title (partial match supported)
//   - capture_region: Captures a region from the virtual screen
//   - list_vision_providers: Lists configured AI vision providers
//   - analyze_image: Analyzes an image with a custom prompt
//   - extract_text: Extracts text from an image as formatted markdown
//   - find_region: Finds bounding box coordinates of a described element
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
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

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
	Version  bool   `short:"v" long:"version" description:"Show version"`
	Help     bool   `short:"h" long:"help" description:"Show help"`
	Config   string `long:"config" description:"Path to config file"`
	LogLevel string `short:"l" long:"log-level" description:"Log level" default:"info"`
	Color    string `long:"color" description:"Color output: always|never|auto" default:"auto"`
	Listen   string `long:"listen" description:"Listen on TCP address (e.g. 127.0.0.1:11777) or 'stdio' for stdio mode" default:""`
	Stdio    bool   `long:"stdio" description:"Run in stdio mode (overrides --listen)"`
}

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
				fmt.Fprintln(os.Stderr, "Error: command required")
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
	if opts.Color != "auto" {
		cfg.Color = opts.Color
	}
	if opts.Listen != "" {
		cfg.Listen = opts.Listen
	}

	logging.Init(cfg.LogLevel, cfg.Color)

	if opts.Help {
		parser.WriteHelp(os.Stdout)
		os.Exit(0)
	}

	if opts.Version {
		fmt.Println("screenshot-mcp-server version 0.1.0")
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

	server := mcp.NewServer(&mcp.Implementation{
		Name:    "screenshooter-mcp",
		Version: "0.1.0",
	}, nil)

	registerTools(server, serverTools)

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
		Version: "0.1.0",
	}, nil)

	serverTools := tools.NewTools(capt)

	visionMgr, err := vision.NewManager(cfg.Vision)
	if err != nil {
		logging.Warn().Err(err).Msg("Failed to initialize vision providers")
	} else if visionMgr != nil {
		serverTools.SetVisionManager(visionMgr)
	}

	registerTools(server, serverTools)

	handler := mcp.NewStreamableHTTPHandler(func(req *http.Request) *mcp.Server {
		return server
	}, nil)

	logging.Info().Str("listen", opts.Listen).Msg("HTTP server listening")
	return http.ListenAndServe(opts.Listen, handler)
}

// listMonitorsInput defines the input parameters for the list_monitors MCP tool.
//
// This struct is intentionally empty because list_monitors takes no parameters.
// It exists as a placeholder for the MCP tool schema definition.
type listMonitorsInput struct{}

// captureScreenInput defines the input parameters for the capture_screen MCP tool.
//
// The monitor field is optional. If specified, it identifies which monitor
// to capture. If omitted or empty, the entire virtual screen (all monitors)
// is captured.
//
// The value can be:
//   - A monitor name (e.g., "DP-1" from X11 RANDR)
//   - A monitor alias (e.g., "1", "primary", "middle-1920x1080")
//   - An empty string to capture all screens
//
// When matching aliases, the server performs case-insensitive comparison.
// If no monitor matches the specified value, an error is returned.
type captureScreenInput struct {
	Monitor string `json:"monitor,omitempty" jsonschema:"optional monitor name or alias; captures all if not specified"`
}

// captureWindowInput defines the input parameters for the capture_window MCP tool.
//
// The title field specifies the window to capture. The match is performed using
// case-insensitive substring matching - if the title contains the specified
// string, the window is considered a match.
//
// If multiple windows match the specified title, an error is returned to
// prevent ambiguity. In this case, specify a more unique title string.
//
// If no window matches the specified title, an error is returned.
type captureWindowInput struct {
	Title string `json:"title" jsonschema:"window title to capture (partial match supported)"`
}

// captureRegionInput defines the input parameters for the capture_region MCP tool.
//
// The x and y fields specify the coordinates of the top-left corner of the
// region to capture, relative to the virtual screen origin (0, 0).
//
// The width and height fields specify the dimensions of the region to capture.
// If the specified region extends beyond the virtual screen bounds, it is clipped
// to the screen boundaries.
//
// Coordinates follow the standard display coordinate system where (0, 0) is
// the top-left corner of the primary monitor. X increases to the right, Y increases
// downward.
type captureRegionInput struct {
	X      int `json:"x" jsonschema:"X coordinate of the top-left corner"`
	Y      int `json:"y" jsonschema:"Y coordinate of the top-left corner"`
	Width  int `json:"width" jsonschema:"width of the region"`
	Height int `json:"height" jsonschema:"height of the region"`
}

// analyzeImageInput defines the input parameters for the analyze_image MCP tool.
type analyzeImageInput struct {
	ImageBase64 string `json:"image_base64" jsonschema:"base64-encoded PNG image data"`
	Prompt      string `json:"prompt" jsonschema:"text prompt describing what analysis to perform"`
	Provider    string `json:"provider,omitempty" jsonschema:"optional provider name; uses default if not specified"`
}

// extractTextInput defines the input parameters for the extract_text MCP tool.
type extractTextInput struct {
	ImageBase64 string `json:"image_base64" jsonschema:"base64-encoded PNG image data"`
	Provider    string `json:"provider,omitempty" jsonschema:"optional provider name; uses default if not specified"`
}

// findRegionInput defines the input parameters for the find_region MCP tool.
type findRegionInput struct {
	ImageBase64 string `json:"image_base64" jsonschema:"base64-encoded PNG image data"`
	Description string `json:"description" jsonschema:"description of the element to find"`
	Provider    string `json:"provider,omitempty" jsonschema:"optional provider name; uses default if not specified"`
}

// registerTools registers all MCP tools with the server.
//
// This function creates and registers MCP tools with the MCP server:
//  1. list_monitors - Lists all available monitors with their names and aliases
//  2. list_windows - Lists all open windows with their titles and IDs
//  3. capture_screen - Captures the full screen or a specific monitor
//  4. capture_window - Captures a specific window by its title
//  5. capture_region - Captures a region from the virtual screen
//  6. list_vision_providers - Lists configured vision providers
//  7. analyze_image - Analyzes an image with a custom prompt
//  8. extract_text - Extracts text from an image as markdown
//  9. find_region - Finds bounding box coordinates of a described element
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
func registerTools(server *mcp.Server, t *tools.Tools) {
	toolNames := []string{}

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_monitors",
		Description: "List all available monitors with their names and aliases",
	}, func(ctx context.Context, req *mcp.CallToolRequest, _ *listMonitorsInput) (*mcp.CallToolResult, any, error) {
		logging.Debug().Str("tool", "list_monitors").Msg("Tool called")
		monitors, err := t.ListMonitors(ctx)
		if err != nil {
			logging.Error().Err(err).Str("tool", "list_monitors").Msg("Tool failed")
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("Failed to list monitors: %v", err)},
				},
				IsError: true,
			}, nil, nil
		}
		logging.Debug().Int("count", len(monitors)).Msg("Monitors listed")

		jsonData, err := json.Marshal(monitors)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("Failed to marshal monitors: %v", err)},
				},
				IsError: true,
			}, nil, nil
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: string(jsonData)},
			},
		}, nil, nil
	})
	toolNames = append(toolNames, "list_monitors")

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_windows",
		Description: "List all open windows with their titles and IDs",
	}, func(ctx context.Context, req *mcp.CallToolRequest, _ *listMonitorsInput) (*mcp.CallToolResult, any, error) {
		logging.Debug().Str("tool", "list_windows").Msg("Tool called")
		windows, err := t.ListWindows(ctx)
		if err != nil {
			logging.Error().Err(err).Str("tool", "list_windows").Msg("Tool failed")
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("Failed to list windows: %v", err)},
				},
				IsError: true,
			}, nil, nil
		}
		logging.Debug().Int("count", len(windows)).Msg("Windows listed")

		jsonData, err := json.Marshal(windows)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("Failed to marshal windows: %v", err)},
				},
				IsError: true,
			}, nil, nil
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: string(jsonData)},
			},
		}, nil, nil
	})
	toolNames = append(toolNames, "list_windows")

	mcp.AddTool(server, &mcp.Tool{
		Name:        "capture_screen",
		Description: "Capture the full screen or a specific monitor",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args *captureScreenInput) (*mcp.CallToolResult, any, error) {
		logging.Debug().Str("tool", "capture_screen").Str("monitor", args.Monitor).Msg("Tool called")
		imgData, err := t.CaptureScreen(ctx, args.Monitor)
		if err != nil {
			logging.Error().Err(err).Str("tool", "capture_screen").Msg("Tool failed")
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("Failed to capture screen: %v", err)},
				},
				IsError: true,
			}, nil, nil
		}
		logging.Debug().Int("size", len(imgData)).Msg("Screen captured")
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.ImageContent{Data: imgData, MIMEType: "image/png"},
			},
		}, nil, nil
	})
	toolNames = append(toolNames, "capture_screen")

	mcp.AddTool(server, &mcp.Tool{
		Name:        "capture_window",
		Description: "Capture a specific window by its title (partial match supported)",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args *captureWindowInput) (*mcp.CallToolResult, any, error) {
		logging.Debug().Str("tool", "capture_window").Str("title", args.Title).Msg("Tool called")
		imgData, err := t.CaptureWindow(ctx, args.Title)
		if err != nil {
			logging.Error().Err(err).Str("tool", "capture_window").Msg("Tool failed")
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("Failed to capture window: %v", err)},
				},
				IsError: true,
			}, nil, nil
		}
		logging.Debug().Int("size", len(imgData)).Msg("Window captured")
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.ImageContent{Data: imgData, MIMEType: "image/png"},
			},
		}, nil, nil
	})
	toolNames = append(toolNames, "capture_window")

	mcp.AddTool(server, &mcp.Tool{
		Name:        "capture_region",
		Description: "Capture a region from the virtual screen",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args *captureRegionInput) (*mcp.CallToolResult, any, error) {
		logging.Debug().Str("tool", "capture_region").Int("x", args.X).Int("y", args.Y).Int("width", args.Width).Int("height", args.Height).Msg("Tool called")
		imgData, err := t.CaptureRegion(ctx, args.X, args.Y, args.Width, args.Height)
		if err != nil {
			logging.Error().Err(err).Str("tool", "capture_region").Msg("Tool failed")
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("Failed to capture region: %v", err)},
				},
				IsError: true,
			}, nil, nil
		}
		logging.Debug().Int("size", len(imgData)).Msg("Region captured")
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.ImageContent{Data: imgData, MIMEType: "image/png"},
			},
		}, nil, nil
	})
	toolNames = append(toolNames, "capture_region")

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_vision_providers",
		Description: "List all configured AI vision providers",
	}, func(ctx context.Context, req *mcp.CallToolRequest, _ *listMonitorsInput) (*mcp.CallToolResult, any, error) {
		logging.Debug().Str("tool", "list_vision_providers").Msg("Tool called")
		providers, err := t.ListVisionProviders(ctx)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("Failed to list vision providers: %v", err)},
				},
				IsError: true,
			}, nil, nil
		}

		jsonData, err := json.Marshal(providers)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("Failed to marshal providers: %v", err)},
				},
				IsError: true,
			}, nil, nil
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: string(jsonData)},
			},
		}, nil, nil
	})
	toolNames = append(toolNames, "list_vision_providers")

	mcp.AddTool(server, &mcp.Tool{
		Name:        "analyze_image",
		Description: "Analyze an image using AI vision providers",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args *analyzeImageInput) (*mcp.CallToolResult, any, error) {
		logging.Debug().Str("tool", "analyze_image").Str("provider", args.Provider).Msg("Tool called")
		imageData, err := base64.StdEncoding.DecodeString(args.ImageBase64)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("Failed to decode image: %v", err)},
				},
				IsError: true,
			}, nil, nil
		}

		result, err := t.AnalyzeImage(ctx, imageData, args.Prompt, args.Provider)
		if err != nil {
			logging.Error().Err(err).Str("tool", "analyze_image").Msg("Tool failed")
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("Failed to analyze image: %v", err)},
				},
				IsError: true,
			}, nil, nil
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: result},
			},
		}, nil, nil
	})
	toolNames = append(toolNames, "analyze_image")

	mcp.AddTool(server, &mcp.Tool{
		Name:        "extract_text",
		Description: "Extract text from an image as formatted markdown",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args *extractTextInput) (*mcp.CallToolResult, any, error) {
		logging.Debug().Str("tool", "extract_text").Str("provider", args.Provider).Msg("Tool called")
		imageData, err := base64.StdEncoding.DecodeString(args.ImageBase64)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("Failed to decode image: %v", err)},
				},
				IsError: true,
			}, nil, nil
		}

		result, err := t.ExtractText(ctx, imageData, args.Provider)
		if err != nil {
			logging.Error().Err(err).Str("tool", "extract_text").Msg("Tool failed")
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("Failed to extract text: %v", err)},
				},
				IsError: true,
			}, nil, nil
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: result},
			},
		}, nil, nil
	})
	toolNames = append(toolNames, "extract_text")

	mcp.AddTool(server, &mcp.Tool{
		Name:        "find_region",
		Description: "Find bounding box coordinates of a described element in an image",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args *findRegionInput) (*mcp.CallToolResult, any, error) {
		logging.Debug().Str("tool", "find_region").Str("provider", args.Provider).Msg("Tool called")
		imageData, err := base64.StdEncoding.DecodeString(args.ImageBase64)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("Failed to decode image: %v", err)},
				},
				IsError: true,
			}, nil, nil
		}

		result, err := t.FindRegion(ctx, imageData, args.Description, args.Provider)
		if err != nil {
			logging.Error().Err(err).Str("tool", "find_region").Msg("Tool failed")
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("Failed to find region: %v", err)},
				},
				IsError: true,
			}, nil, nil
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: result},
			},
		}, nil, nil
	})
	toolNames = append(toolNames, "find_region")

	logging.Info().Strs("tools", toolNames).Msg("Tools registered")
}
