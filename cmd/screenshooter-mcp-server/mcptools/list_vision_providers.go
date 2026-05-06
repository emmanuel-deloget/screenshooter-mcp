package mcptools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/emmanuel-deloget/screenshooter-mcp/internal/logging"
	"github.com/emmanuel-deloget/screenshooter-mcp/internal/tools"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func ListVisionProvides(t *tools.Tools) func(ctx context.Context, req *mcp.CallToolRequest, _ *EmptyArgs) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, _ *EmptyArgs) (*mcp.CallToolResult, any, error) {
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
	}
}
