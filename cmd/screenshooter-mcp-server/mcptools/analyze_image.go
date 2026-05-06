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

// AnalyzeImageInput defines the input parameters for the analyze_image MCP tool.
type AnalyzeImageInput struct {
	ImageBase64 string `json:"image_base64" jsonschema:"base64-encoded PNG image data"`
	Prompt      string `json:"prompt" jsonschema:"text prompt describing what analysis to perform"`
	Provider    string `json:"provider,omitempty" jsonschema:"optional provider name; uses default if not specified"`
	Timeout     int    `json:"timeout,omitempty" jsonschema:"optional timeout in seconds; 0 uses provider default"`
}

func AnalyzeImage(checkAccess utils.AccessCheckerFn, t *tools.Tools) func(ctx context.Context, req *mcp.CallToolRequest, args *AnalyzeImageInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, args *AnalyzeImageInput) (*mcp.CallToolResult, any, error) {
		logging.Debug().Str("tool", "analyze_image").Str("provider", args.Provider).Msg("Tool called")
		if result := checkAccess("analyze_image"); result != nil {
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
		result, err := t.AnalyzeImage(ctx, imageData, args.Prompt, args.Provider, args.Timeout)
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
	}
}
