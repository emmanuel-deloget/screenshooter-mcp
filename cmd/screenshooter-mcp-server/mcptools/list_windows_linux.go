package mcptools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/emmanuel-deloget/screenshooter-mcp/cmd/screenshooter-mcp-server/utils"
	"github.com/emmanuel-deloget/screenshooter-mcp/internal/logging"
	"github.com/emmanuel-deloget/screenshooter-mcp/internal/tools"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func ListWindows(checkAccess utils.AccessCheckerFn, t *tools.Tools) func(ctx context.Context, req *mcp.CallToolRequest, _ *EmptyArgs) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, _ *EmptyArgs) (*mcp.CallToolResult, any, error) {
		logging.Debug().Str("tool", "list_windows").Msg("Tool called")
		if result := checkAccess("list_windows"); result != nil {
			return result, nil, nil
		}
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
	}
}
