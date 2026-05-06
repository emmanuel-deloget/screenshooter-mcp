package mcptools

import (
	"context"
	"fmt"
	"time"

	"github.com/emmanuel-deloget/screenshooter-mcp/cmd/screenshooter-mcp-server/access"
	"github.com/emmanuel-deloget/screenshooter-mcp/internal/logging"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// AllowToolAccessInput defines the input for the allow_tool_access MCP tool.
type AllowToolAccessInput struct {
	Tool     string `json:"tool" jsonschema:"tool name to grant temporary access to"`
	Duration int    `json:"duration,omitempty" jsonschema:"duration in seconds (default: config default)"`
}

func AllowToolAccess(am *access.AccessManager) func(ctx context.Context, req *mcp.CallToolRequest, args *AllowToolAccessInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, args *AllowToolAccessInput) (*mcp.CallToolResult, any, error) {
		logging.Debug().Str("tool", "allow-tool-access").Str("target", args.Tool).Msg("Tool called")
		if !am.IsToolRegistered(args.Tool) {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("unknown tool: %q", args.Tool)},
				},
				IsError: true,
			}, nil, nil
		}
		duration := time.Duration(args.Duration) * time.Second
		am.GrantTemporary(args.Tool, duration)
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("temporary access granted for tool %q", args.Tool)},
			},
		}, nil, nil
	}
}
