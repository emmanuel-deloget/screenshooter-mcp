package mcptools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/emmanuel-deloget/screenshooter-mcp/internal/logging"
	"github.com/emmanuel-deloget/screenshooter-mcp/internal/tools"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func ListMonitors(t *tools.Tools) func(ctx context.Context, req *mcp.CallToolRequest, _ *EmptyArgs) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, _ *EmptyArgs) (*mcp.CallToolResult, any, error) {
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
	}
}
