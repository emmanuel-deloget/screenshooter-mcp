package mcptools

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/emmanuel-deloget/screenshooter-mcp/cmd/screenshooter-mcp-server/utils"
	"github.com/emmanuel-deloget/screenshooter-mcp/internal/logging"
	"github.com/emmanuel-deloget/screenshooter-mcp/internal/tools"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ExtractTextInput defines the input parameters for the extract_text MCP tool.
type ExtractTextInput struct {
	ImageBase64 string `json:"image_base64" jsonschema:"base64-encoded PNG image data"`
	Provider    string `json:"provider,omitempty" jsonschema:"optional provider name; uses default if not specified"`
	Timeout     int    `json:"timeout,omitempty" jsonschema:"optional timeout in seconds; 0 uses provider default"`
}

func ExtractText(checkAccess utils.AccessCheckerFn, t *tools.Tools) func(ctx context.Context, req *mcp.CallToolRequest, args *ExtractTextInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, args *ExtractTextInput) (*mcp.CallToolResult, any, error) {
		logging.Debug().Str("tool", "extract_text").Str("provider", args.Provider).Msg("Tool called")
		if result := checkAccess("extract_text"); result != nil {
			return result, nil, nil
		}
		imageData, err := base64.StdEncoding.DecodeString(args.ImageBase64)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("Failed to decode image: %v", err)},
				},
				IsError: true,
			}, nil, nil
		}
		result, err := t.ExtractText(ctx, imageData, args.Provider, args.Timeout)
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
	}
}
