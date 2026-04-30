package main

import (
	"reflect"
	"testing"
)

func TestParseRegionResponse(t *testing.T) {
	tests := []struct {
		name     string
		response string
		want     RegionResult
	}{
		{
			name:     "clean JSON",
			response: `{"x": 100, "y": 200, "width": 300, "height": 400}`,
			want:     RegionResult{X: 100, Y: 200, Width: 300, Height: 400},
		},
		{
			name:     "JSON with extra whitespace",
			response: `  {  "x": 10, "y": 20, "width": 50, "height": 60  }  `,
			want:     RegionResult{X: 10, Y: 20, Width: 50, Height: 60},
		},
		{
			name:     "markdown code block",
			response: "```json\n{\"x\": 1, \"y\": 2, \"width\": 3, \"height\": 4}\n```",
			want:     RegionResult{X: 1, Y: 2, Width: 3, Height: 4},
		},
		{
			name:     "markdown without json tag",
			response: "```\n{\"x\": 5, \"y\": 6, \"width\": 7, \"height\": 8}\n```",
			want:     RegionResult{X: 5, Y: 6, Width: 7, Height: 8},
		},
		{
			name:     "text with JSON embedded",
			response: "The coordinates are:\n{\"x\": 10, \"y\": 20, \"width\": 30, \"height\": 40}\nHope this helps!",
			want:     RegionResult{X: 10, Y: 20, Width: 30, Height: 40},
		},
		{
			name:     "numbers only fallback",
			response: "The button is at 100, 200 with size 50x30",
			want:     RegionResult{X: 100, Y: 200, Width: 50, Height: 30},
		},
		{
			name:     "empty response",
			response: "",
			want:     RegionResult{},
		},
		{
			name:     "no numbers",
			response: "I cannot find that element",
			want:     RegionResult{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseRegionResponse(tt.response)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseRegionResponse() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestFindJSONBlock(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want int
	}{
		{"json marker", "```json\n{...}", 8},
		{"code block marker", "```\n{...}", 4},
		{"direct brace", "{...}", 0},
		{"no json", "hello world", -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := findJSONBlock(tt.s)
			if got != tt.want {
				t.Errorf("findJSONBlock() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestFindJSONEnd(t *testing.T) {
	tests := []struct {
		name  string
		s     string
		start int
		want  int
	}{
		{"simple object", `{"x": 1}`, 0, 8},
		{"nested object", `{"a": {"b": 1}}`, 0, 15},
		{"with prefix", `text{"x": 1}more`, 4, 12},
		{"no close", `{"x": 1`, 0, -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := findJSONEnd(tt.s, tt.start)
			if got != tt.want {
				t.Errorf("findJSONEnd() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestParseRegionNumbers(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want RegionResult
	}{
		{"four numbers", "100 200 300 400", RegionResult{X: 100, Y: 200, Width: 300, Height: 400}},
		{"comma separated", "10, 20, 30, 40", RegionResult{X: 10, Y: 20, Width: 30, Height: 40}},
		{"less than four", "100 200", RegionResult{}},
		{"more than four", "1 2 3 4 5 6", RegionResult{X: 1, Y: 2, Width: 3, Height: 4}},
		{"mixed text", "x=10 y=20 w=30 h=40", RegionResult{X: 10, Y: 20, Width: 30, Height: 40}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseRegionNumbers(tt.s)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseRegionNumbers() = %+v, want %+v", got, tt.want)
			}
		})
	}
}
