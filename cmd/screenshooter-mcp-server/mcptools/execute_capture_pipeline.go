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

// ExecutePipelineInput defines the input parameters for the execute_capture_pipeline MCP tool.
type ExecutePipelineInput struct {
	Pipeline any `json:"pipeline" jsonschema:"ordered list of pipeline steps to execute"`
}

func ExecuteCapturePipeline(checkAccess utils.AccessCheckerFn, t *tools.Tools) func(ctx context.Context, req *mcp.CallToolRequest, args *ExecutePipelineInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, args *ExecutePipelineInput) (*mcp.CallToolResult, any, error) {
		steps, err := utils.DeserializeSlice[tools.PipelineStep](args.Pipeline)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("Invalid pipeline argument: %v", err)},
				},
				IsError: true,
			}, nil, nil
		}
		logging.Debug().Int("steps", len(steps)).Str("tool", "execute_capture_pipeline").Msg("Tool called")
		if result := checkAccess("execute_capture_pipeline"); result != nil {
			return result, nil, nil
		}
		imgBase64, text, err := tools.ExecutePipeline(ctx, steps, t)
		if err != nil {
			logging.Error().Err(err).Str("tool", "execute_capture_pipeline").Msg("Tool failed")
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("Pipeline execution failed: %v", err)},
				},
				IsError: true,
			}, nil, nil
		}
		var content []mcp.Content
		if imgBase64 != "" {
			imgData, err := base64.StdEncoding.DecodeString(imgBase64)
			if err != nil {
				return &mcp.CallToolResult{
					Content: []mcp.Content{
						&mcp.TextContent{Text: fmt.Sprintf("Failed to decode pipeline image result: %v", err)},
					},
					IsError: true,
				}, nil, nil
			}
			content = append(content, &mcp.ImageContent{Data: imgData, MIMEType: "image/png"})
		}
		if text != "" {
			content = append(content, &mcp.TextContent{Text: text})
		}
		return &mcp.CallToolResult{
			Content: content,
		}, nil, nil
	}
}
