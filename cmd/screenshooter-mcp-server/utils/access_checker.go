package utils

import (
	"fmt"

	"github.com/emmanuel-deloget/screenshooter-mcp/cmd/screenshooter-mcp-server/access"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type AccessCheckerFn func(tool string) *mcp.CallToolResult

// CheckAccess returns an error CallToolResult if access is denied or requires asking.
// Returns nil if access is allowed.
func CheckAccess(tool string, am *access.AccessManager) *mcp.CallToolResult {
	switch am.GetAccess(tool) {
	case "deny":
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("access to tool %q is denied", tool)},
			},
			IsError: true,
		}
	case "ask":
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("access to tool %q requires user permission; call 'allow_tool_access' to grant temporary access", tool)},
			},
			IsError: true,
		}
	}
	return nil
}
