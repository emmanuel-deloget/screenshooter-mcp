package mcptools

import (
	"context"

	"github.com/emmanuel-deloget/screenshooter-mcp/internal/logging"
	"github.com/emmanuel-deloget/screenshooter-mcp/internal/tools"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func GetSkillInfoForAgent(t *tools.Tools) func(ctx context.Context, req *mcp.CallToolRequest, _ *EmptyArgs) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, _ *EmptyArgs) (*mcp.CallToolResult, any, error) {
		logging.Debug().Str("tool", "get_skill_info_for_agent").Msg("Tool called")
		skill := t.GetSkillInfo()
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: skill},
			},
		}, nil, nil
	}
}
