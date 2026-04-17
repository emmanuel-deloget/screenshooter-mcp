// Copyright 2025 Emmanuel Deloget. All rights reserved.
// Use of this source code is governed by the license that can be found in the LICENSE file.

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/emmanuel-deloget/screenshooter-mcp/internal/capture"
	"github.com/emmanuel-deloget/screenshooter-mcp/internal/capture/wayland"
	"github.com/emmanuel-deloget/screenshooter-mcp/internal/capture/x11"
	"github.com/emmanuel-deloget/screenshooter-mcp/internal/logging"
	"github.com/emmanuel-deloget/screenshooter-mcp/internal/tools"
	"github.com/emmanuel-deloget/screenshooter-mcp/internal/vision"
	"github.com/jessevdk/go-flags"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type Options struct {
	Version          bool                 `short:"v" long:"version" description:"Show version"`
	Help             bool                 `short:"h" long:"help" description:"Show help"`
	VisionModel      string               `short:"m" long:"vision-model" description:"Vision model name" default:"qwen3-vl:4b"`
	ListVisionModels bool                 `long:"list-vision-models" description:"List available vision models"`
	VisionQuality    vision.VisionQuality `short:"q" long:"vision-quality" description:"Vision quality" default:"middle"`
	LogLevel         string               `short:"l" long:"log-level" description:"Log level" default:"info"`
	Color            string               `long:"color" description:"Color output: always|never|auto" default:"auto"`
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

	logging.Init(opts.LogLevel, opts.Color)

	if opts.Help {
		parser.WriteHelp(os.Stdout)
		os.Exit(0)
	}

	if opts.Version {
		fmt.Println("screenshot-mcp-server version 0.1.0")
		os.Exit(0)
	}

	if opts.ListVisionModels {
		listVisionModels()
		os.Exit(0)
	}

	if err := run(&opts); err != nil {
		logging.Error().Err(err).Msg("Server error")
		os.Exit(1)
	}
}

func listVisionModels() {
	fmt.Println("Available vision models:")
	fmt.Println("  qwen3-vl:2b  - Fast, CPU-efficient (1.9GB)")
	fmt.Println("  qwen3-vl:4b  - Balanced (3.3GB)")
	fmt.Println("  qwen3-vl:8b  - Higher quality (6.1GB)")
}

func run(opts *Options) error {
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
		capt = x11.NewX11Capture()
	case capture.EnvironmentWayland:
		logging.Debug().Msg("Creating Wayland capture")
		capt = wayland.NewWaylandCapture()
	default:
		return fmt.Errorf("unsupported environment: %s", env)
	}

	logging.Info().Str("model", opts.VisionModel).Str("quality", string(opts.VisionQuality)).Msg("Starting Ollama vision manager")
	visionMgr, err := vision.NewManager(
		vision.WithModel(opts.VisionModel),
		vision.WithQuality(opts.VisionQuality),
	)
	if err != nil {
		logging.Error().Err(err).Msg("Failed to start vision manager")
		return fmt.Errorf("failed to start vision manager: %w", err)
	}
	logging.Info().Str("url", visionMgr.URL()).Int("pid", visionMgr.PID()).Msg("Ollama running")

	defer visionMgr.Stop()

	serverTools := tools.NewTools(capt, visionMgr)

	server := mcp.NewServer(&mcp.Implementation{
		Name:    "screenshooter-mcp",
		Version: "0.1.0",
	}, nil)

	registerTools(server, serverTools)

	logging.Info().Msg("MCP server running on stdio")
	return server.Run(context.Background(), &mcp.StdioTransport{})
}

type listMonitorsInput struct{}

type captureScreenInput struct {
	Monitor string `json:"monitor,omitempty" jsonschema:"optional monitor name or alias; captures all if not specified"`
}

type captureWindowInput struct {
	WindowID int64 `json:"window_id" jsonschema:"the window ID to capture"`
}

type captureRegionInput struct {
	X      int `json:"x" jsonschema:"X coordinate of the top-left corner"`
	Y      int `json:"y" jsonschema:"Y coordinate of the top-left corner"`
	Width  int `json:"width" jsonschema:"width of the region"`
	Height int `json:"height" jsonschema:"height of the region"`
}

type findElementInput struct {
	Image       string `json:"image" jsonschema:"base64-encoded PNG image"`
	Description string `json:"description" jsonschema:"natural language description of the element to find"`
}

type captureElementInput struct {
	Image       string `json:"image" jsonschema:"base64-encoded PNG image"`
	Description string `json:"description" jsonschema:"natural language description of the element to find"`
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
		Name:        "capture_screen",
		Description: "Capture the full screen or a specific monitor",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args *captureScreenInput) (*mcp.CallToolResult, any, error) {
		logging.Debug().Str("tool", "capture_screen").Str("monitor", args.Monitor).Msg("Tool called")
		imgBase64, err := t.CaptureScreen(ctx, args.Monitor)
		if err != nil {
			logging.Error().Err(err).Str("tool", "capture_screen").Msg("Tool failed")
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("Failed to capture screen: %v", err)},
				},
				IsError: true,
			}, nil, nil
		}
		logging.Debug().Int("size", len(imgBase64)).Msg("Screen captured")
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: imgBase64},
			},
		}, nil, nil
	})
	toolNames = append(toolNames, "capture_screen")

	mcp.AddTool(server, &mcp.Tool{
		Name:        "capture_window",
		Description: "Capture a specific window by its ID",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args *captureWindowInput) (*mcp.CallToolResult, any, error) {
		logging.Debug().Str("tool", "capture_window").Int64("window_id", args.WindowID).Msg("Tool called")
		imgBase64, err := t.CaptureWindow(ctx, args.WindowID)
		if err != nil {
			logging.Error().Err(err).Str("tool", "capture_window").Msg("Tool failed")
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("Failed to capture window: %v", err)},
				},
				IsError: true,
			}, nil, nil
		}
		logging.Debug().Int("size", len(imgBase64)).Msg("Window captured")
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: imgBase64},
			},
		}, nil, nil
	})
	toolNames = append(toolNames, "capture_window")

	mcp.AddTool(server, &mcp.Tool{
		Name:        "capture_region",
		Description: "Capture a region from the virtual screen",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args *captureRegionInput) (*mcp.CallToolResult, any, error) {
		logging.Debug().Str("tool", "capture_region").Int("x", args.X).Int("y", args.Y).Int("width", args.Width).Int("height", args.Height).Msg("Tool called")
		imgBase64, err := t.CaptureRegion(ctx, args.X, args.Y, args.Width, args.Height)
		if err != nil {
			logging.Error().Err(err).Str("tool", "capture_region").Msg("Tool failed")
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("Failed to capture region: %v", err)},
				},
				IsError: true,
			}, nil, nil
		}
		logging.Debug().Int("size", len(imgBase64)).Msg("Region captured")
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: imgBase64},
			},
		}, nil, nil
	})
	toolNames = append(toolNames, "capture_region")

	mcp.AddTool(server, &mcp.Tool{
		Name:        "find_element",
		Description: "Find an element in a screenshot using vision model",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args *findElementInput) (*mcp.CallToolResult, any, error) {
		logging.Debug().Str("tool", "find_element").Int("image_size", len(args.Image)).Str("description", args.Description).Msg("Tool called")
		element, err := t.FindElement(ctx, args.Image, args.Description)
		if err != nil {
			logging.Error().Err(err).Str("tool", "find_element").Msg("Tool failed")
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("Failed to find element: %v", err)},
				},
				IsError: true,
			}, nil, nil
		}
		logging.Debug().Interface("bbox", element.BoundingBox).Float64("confidence", element.Confidence).Msg("Element found")

		jsonData, err := json.Marshal(element)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("Failed to marshal element: %v", err)},
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
	toolNames = append(toolNames, "find_element")

	mcp.AddTool(server, &mcp.Tool{
		Name:        "capture_element",
		Description: "Find an element in a screenshot and return cropped image",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args *captureElementInput) (*mcp.CallToolResult, any, error) {
		logging.Debug().Str("tool", "capture_element").Int("image_size", len(args.Image)).Str("description", args.Description).Msg("Tool called")
		imgBase64, err := t.CaptureElement(ctx, args.Image, args.Description)
		if err != nil {
			logging.Error().Err(err).Str("tool", "capture_element").Msg("Tool failed")
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("Failed to capture element: %v", err)},
				},
				IsError: true,
			}, nil, nil
		}
		logging.Debug().Int("size", len(imgBase64)).Msg("Element captured")
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: imgBase64},
			},
		}, nil, nil
	})
	toolNames = append(toolNames, "capture_element")

	logging.Info().Strs("tools", toolNames).Msg("Tools registered")
}
