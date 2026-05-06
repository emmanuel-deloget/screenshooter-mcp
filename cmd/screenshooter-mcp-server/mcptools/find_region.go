package mcptools

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/emmanuel-deloget/screenshooter-mcp/cmd/screenshooter-mcp-server/utils"
	"github.com/emmanuel-deloget/screenshooter-mcp/internal/logging"
	"github.com/emmanuel-deloget/screenshooter-mcp/internal/tools"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// FindRegionInput defines the input parameters for the find_region MCP tool.
type FindRegionInput struct {
	ImageBase64 string `json:"image_base64" jsonschema:"base64-encoded PNG image data"`
	Description string `json:"description" jsonschema:"description of the element to find"`
	Provider    string `json:"provider,omitempty" jsonschema:"optional provider name; uses default if not specified"`
	Timeout     int    `json:"timeout,omitempty" jsonschema:"optional timeout in seconds; 0 uses provider default"`
}

// RegionResult represents the bounding box coordinates returned by find_region.
type RegionResult struct {
	X      int `json:"x"`
	Y      int `json:"y"`
	Width  int `json:"width"`
	Height int `json:"height"`
}

func FindRegion(checkAccess utils.AccessCheckerFn, t *tools.Tools) func(ctx context.Context, req *mcp.CallToolRequest, args *FindRegionInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, args *FindRegionInput) (*mcp.CallToolResult, any, error) {
		logging.Debug().Str("tool", "find_region").Str("provider", args.Provider).Msg("Tool called")
		if result := checkAccess("find_region"); result != nil {
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
		result, err := t.FindRegion(ctx, imageData, args.Description, args.Provider, args.Timeout)
		if err != nil {
			logging.Error().Err(err).Str("tool", "find_region").Msg("Tool failed")
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("Failed to find region: %v", err)},
				},
				IsError: true,
			}, nil, nil
		}
		region := parseRegionResponse(result)
		jsonData, err := json.Marshal(region)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("Failed to parse region result: %v", err)},
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

// parseRegionResponse extracts the region coordinates from the AI model response.
// It attempts to parse the response as JSON {x, y, width, height}.
func parseRegionResponse(response string) RegionResult {
	var region RegionResult
	if err := json.Unmarshal([]byte(response), &region); err == nil {
		return region
	}

	// Try to extract JSON from markdown code blocks
	if start := findJSONBlock(response); start >= 0 {
		end := findJSONEnd(response, start)
		if end > start {
			if err := json.Unmarshal([]byte(response[start:end]), &region); err == nil {
				return region
			}
		}
	}

	// Fallback: try to find numbers in the response
	region = parseRegionNumbers(response)
	return region
}

// findJSONBlock finds the start of a JSON block in markdown.
func findJSONBlock(s string) int {
	markers := []string{"```json\n", "```\n", "{"}
	for _, m := range markers {
		if idx := index(s, m); idx >= 0 {
			if m == "{" {
				return idx
			}
			return idx + len(m)
		}
	}
	return -1
}

// findJSONEnd finds the matching closing brace for a JSON object.
func findJSONEnd(s string, start int) int {
	depth := 0
	for i := start; i < len(s); i++ {
		switch s[i] {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return i + 1
			}
		}
	}
	return -1
}

// index is a simple strings.Index replacement.
func index(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// parseRegionNumbers extracts numbers from text as a last resort fallback.
func parseRegionNumbers(s string) RegionResult {
	var nums []int
	var current int
	inNum := false
	for _, c := range s {
		if c >= '0' && c <= '9' {
			current = current*10 + int(c-'0')
			inNum = true
		} else if inNum {
			nums = append(nums, current)
			current = 0
			inNum = false
			if len(nums) == 4 {
				break
			}
		}
	}
	if inNum && len(nums) < 4 {
		nums = append(nums, current)
	}

	if len(nums) >= 4 {
		return RegionResult{X: nums[0], Y: nums[1], Width: nums[2], Height: nums[3]}
	}
	return RegionResult{}
}
