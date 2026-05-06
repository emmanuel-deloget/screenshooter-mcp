package mcptools

import (
	"context"
	"fmt"

	"github.com/emmanuel-deloget/screenshooter-mcp/cmd/screenshooter-mcp-server/utils"
	"github.com/emmanuel-deloget/screenshooter-mcp/internal/logging"
	"github.com/emmanuel-deloget/screenshooter-mcp/internal/tools"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// CaptureWindowInput defines the input parameters for the capture_window MCP tool.
//
// The title field specifies the window to capture. The match is performed using
// case-insensitive substring matching - if the title contains the specified
// string, the window is considered a match.
//
// If multiple windows match the specified title, an error is returned to
// prevent ambiguity. In this case, specify a more unique title string.
//
// If no window matches the specified title, an error is returned.
type CaptureWindowInput struct {
	Title string `json:"title" jsonschema:"window title to capture (partial match supported)"`
}

func CaptureWindow(checkAccess utils.AccessCheckerFn, t *tools.Tools) func(ctx context.Context, req *mcp.CallToolRequest, args *CaptureWindowInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, args *CaptureWindowInput) (*mcp.CallToolResult, any, error) {
		logging.Debug().Str("tool", "capture_window").Str("title", args.Title).Msg("Tool called")
		if result := checkAccess("capture_window"); result != nil {
			return result, nil, nil
		}
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
	}
}
