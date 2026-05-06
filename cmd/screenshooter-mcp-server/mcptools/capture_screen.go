package mcptools

import (
	"context"
	"fmt"

	"github.com/emmanuel-deloget/screenshooter-mcp/cmd/screenshooter-mcp-server/utils"
	"github.com/emmanuel-deloget/screenshooter-mcp/internal/logging"
	"github.com/emmanuel-deloget/screenshooter-mcp/internal/tools"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// CaptureScreenInput defines the input parameters for the capture_screen MCP tool.
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
type CaptureScreenInput struct {
	Monitor string `json:"monitor,omitempty" jsonschema:"optional monitor name or alias; captures all if not specified"`
}

func CaptureScreen(checkAccess utils.AccessCheckerFn, t *tools.Tools) func(ctx context.Context, req *mcp.CallToolRequest, args *CaptureScreenInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, args *CaptureScreenInput) (*mcp.CallToolResult, any, error) {
		logging.Debug().Str("tool", "capture_screen").Str("monitor", args.Monitor).Msg("Tool called")
		if result := checkAccess("capture_screen"); result != nil {
			return result, nil, nil
		}
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
	}
}
