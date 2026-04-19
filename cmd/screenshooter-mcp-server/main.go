// Copyright 2025 Emmanuel Deloget. All rights reserved.
// Use of this source code is governed by the license that can be found in the LICENSE file.

package main

import (
	"context"
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
	"github.com/jessevdk/go-flags"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

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
		return runHttpBridge(opts)
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

	server := mcp.NewServer(&mcp.Implementation{
		Name:    "screenshooter-mcp",
		Version: "0.1.0",
	}, nil)

	registerTools(server, serverTools)

	logging.Info().Msg("MCP server running on stdio")
	return server.Run(context.Background(), &mcp.StdioTransport{})
}

func runHttpBridge(opts *Options) error {
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
	registerTools(server, serverTools)

	handler := mcp.NewStreamableHTTPHandler(func(req *http.Request) *mcp.Server {
		return server
	}, nil)

	logging.Info().Str("listen", opts.Listen).Msg("HTTP server listening")
	return http.ListenAndServe(opts.Listen, handler)
}

type listMonitorsInput struct{}

type captureScreenInput struct {
	Monitor string `json:"monitor,omitempty" jsonschema:"optional monitor name or alias; captures all if not specified"`
}

type captureWindowInput struct {
	Title string `json:"title" jsonschema:"window title to capture (partial match supported)"`
}

type captureRegionInput struct {
	X      int `json:"x" jsonschema:"X coordinate of the top-left corner"`
	Y      int `json:"y" jsonschema:"Y coordinate of the top-left corner"`
	Width  int `json:"width" jsonschema:"width of the region"`
	Height int `json:"height" jsonschema:"height of the region"`
}

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

	logging.Info().Strs("tools", toolNames).Msg("Tools registered")
}
