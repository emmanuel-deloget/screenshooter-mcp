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

// CompareImagesInput defines the input parameters for the compare_images MCP tool.
type CompareImagesInput struct {
	ImageBase64  string `json:"image_base64" jsonschema:"base64-encoded PNG image data (first image)"`
	Image2Base64 string `json:"image2_base64" jsonschema:"base64-encoded PNG image data (second image)"`
	Prompt       string `json:"prompt,omitempty" jsonschema:"optional comparison prompt; uses default if not specified"`
	Provider     string `json:"provider,omitempty" jsonschema:"optional provider name; uses default if not specified"`
	Timeout      int    `json:"timeout,omitempty" jsonschema:"optional timeout in seconds; 0 uses provider default"`
}

func CompareImages(checkAccess utils.AccessCheckerFn, t *tools.Tools) func(ctx context.Context, req *mcp.CallToolRequest, args *CompareImagesInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, args *CompareImagesInput) (*mcp.CallToolResult, any, error) {
		logging.Debug().Str("tool", "compare_images").Str("provider", args.Provider).Msg("Tool called")
		if result := checkAccess("compare_images"); result != nil {
			return result, nil, nil
		}
		image1Data, err := base64.StdEncoding.DecodeString(args.ImageBase64)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("Failed to decode first image: %v", err)},
				},
				IsError: true,
			}, nil, nil
		}
		image2Data, err := base64.StdEncoding.DecodeString(args.Image2Base64)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("Failed to decode second image: %v", err)},
				},
				IsError: true,
			}, nil, nil
		}
		prompt := args.Prompt
		if prompt == "" {
			prompt = "Describe the differences between these two images. Be specific about what changed."
		}
		result, err := t.CompareImages(ctx, image1Data, image2Data, prompt, args.Provider, args.Timeout)
		if err != nil {
			logging.Error().Err(err).Str("tool", "compare_images").Msg("Tool failed")
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("Failed to compare images: %v", err)},
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
