package main

import (
	"testing"
)

type testStep struct {
	Tool       string         `json:"tool"`
	Parameters map[string]any `json:"parameters,omitempty"`
}

func TestDeserializeSlice(t *testing.T) {
	tests := []struct {
		name    string
		input   any
		wantLen int
		wantErr bool
	}{
		{
			name: "proper array",
			input: []any{
				map[string]any{"tool": "capture_screen", "parameters": map[string]any{}},
				map[string]any{"tool": "capture_window", "parameters": map[string]any{"title": "Files"}},
			},
			wantLen: 2,
			wantErr: false,
		},
		{
			name:    "empty array",
			input:   []any{},
			wantLen: 0,
			wantErr: false,
		},
		{
			name:    "JSON-encoded string",
			input:   `[{"tool":"capture_screen","parameters":{}}]`,
			wantLen: 1,
			wantErr: false,
		},
		{
			name:    "invalid string",
			input:   `not json`,
			wantLen: 0,
			wantErr: true,
		},
		{
			name:    "wrong type",
			input:   42,
			wantLen: 0,
			wantErr: true,
		},
		{
			name:    "nil input",
			input:   nil,
			wantLen: 0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := DeserializeSlice[testStep](tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("DeserializeSlice() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(result) != tt.wantLen {
				t.Errorf("DeserializeSlice() got len %d, want %d", len(result), tt.wantLen)
			}
		})
	}
}
