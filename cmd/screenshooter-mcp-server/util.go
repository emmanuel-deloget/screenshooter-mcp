// Package main provides utility functions for the screenshooter-mcp-server.
package main

import (
	"encoding/json"
	"fmt"
)

// DeserializeSlice handles JSON double-serialization for slice types.
//
// Some MCP clients incorrectly serialize array parameters as JSON-encoded strings
// rather than proper JSON arrays. This function handles both cases:
//
//   - If v is already a slice ([]any), it marshals and unmarshals to convert to []T
//   - If v is a string, it attempts to unmarshal the JSON string into []T
//   - For any other type, it returns an error
//
// This is useful for parameters that contain nested structures like pipelines,
// where the client might send:
//
//	{"pipeline": "[{\"tool\": \"capture_screen\", \"parameters\": {}}]"}  // string (broken)
//	{"pipeline": [{"tool": "capture_screen", "parameters": {}}]}          // array (correct)
func DeserializeSlice[T any](v any) ([]T, error) {
	switch val := v.(type) {
	case []any:
		// Client sent a proper array, but we need to convert []any to []T
		data, err := json.Marshal(val)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize slice: %w", err)
		}
		var result []T
		if err := json.Unmarshal(data, &result); err != nil {
			return nil, fmt.Errorf("failed to deserialize slice: %w", err)
		}
		return result, nil
	case string:
		// Client double-encoded the array as a JSON string
		var result []T
		if err := json.Unmarshal([]byte(val), &result); err != nil {
			return nil, fmt.Errorf("failed to deserialize JSON string: %w", err)
		}
		return result, nil
	default:
		return nil, fmt.Errorf("expected array or JSON-encoded string, got %T", v)
	}
}
