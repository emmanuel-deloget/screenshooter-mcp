package mcptools

import (
	"context"
	"fmt"

	"github.com/emmanuel-deloget/screenshooter-mcp/cmd/screenshooter-mcp-server/utils"
	"github.com/emmanuel-deloget/screenshooter-mcp/internal/logging"
	"github.com/emmanuel-deloget/screenshooter-mcp/internal/tools"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// CaptureRegionInput defines the input parameters for the capture_region MCP tool.
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
type CaptureRegionInput struct {
	X      int `json:"x" jsonschema:"X coordinate of the top-left corner"`
	Y      int `json:"y" jsonschema:"Y coordinate of the top-left corner"`
	Width  int `json:"width" jsonschema:"width of the region"`
	Height int `json:"height" jsonschema:"height of the region"`
}

func CaptureRegion(checkAccess utils.AccessCheckerFn, t *tools.Tools) func(ctx context.Context, req *mcp.CallToolRequest, args *CaptureRegionInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, args *CaptureRegionInput) (*mcp.CallToolResult, any, error) {
		logging.Debug().Str("tool", "capture_region").Int("x", args.X).Int("y", args.Y).Int("width", args.Width).Int("height", args.Height).Msg("Tool called")
		if result := checkAccess("capture_region"); result != nil {
			return result, nil, nil
		}
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
	}
}
