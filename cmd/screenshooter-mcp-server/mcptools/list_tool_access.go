package mcptools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/emmanuel-deloget/screenshooter-mcp/cmd/screenshooter-mcp-server/access"
	"github.com/emmanuel-deloget/screenshooter-mcp/internal/logging"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func ListToolAccess(am *access.AccessManager) func(ctx context.Context, req *mcp.CallToolRequest, _ *EmptyArgs) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, _ *EmptyArgs) (*mcp.CallToolResult, any, error) {
		logging.Debug().Str("tool", "list_tool_access").Msg("Tool called")
		jsonData, err := json.MarshalIndent(am.ListAccess(), "", "  ")
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("Failed to marshal access list: %v", err)},
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
