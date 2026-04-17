package capture

import "testing"

func TestBoundingBoxIsValid(t *testing.T) {
	tests := []struct {
		name     string
		bbox     BoundingBox
		expected bool
	}{
		{
			name:     "valid bounding box",
			bbox:     BoundingBox{X1: 0, Y1: 0, X2: 100, Y2: 100},
			expected: true,
		},
		{
			name:     "valid with negative coordinates",
			bbox:     BoundingBox{X1: -50, Y1: -50, X2: 100, Y2: 100},
			expected: false,
		},
		{
			name:     "zero width",
			bbox:     BoundingBox{X1: 0, Y1: 0, X2: 0, Y2: 100},
			expected: false,
		},
		{
			name:     "zero height",
			bbox:     BoundingBox{X1: 0, Y1: 0, X2: 100, Y2: 0},
			expected: false,
		},
		{
			name:     "inverted coordinates",
			bbox:     BoundingBox{X1: 100, Y1: 100, X2: 0, Y2: 0},
			expected: false,
		},
		{
			name:     "negative x2",
			bbox:     BoundingBox{X1: 0, Y1: 0, X2: -10, Y2: 100},
			expected: false,
		},
		{
			name:     "negative y2",
			bbox:     BoundingBox{X1: 0, Y1: 0, X2: 100, Y2: -10},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.bbox.IsValid()
			if result != tt.expected {
				t.Errorf("BoundingBox.IsValid() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestBoundingBoxWidth(t *testing.T) {
	tests := []struct {
		name     string
		bbox     BoundingBox
		expected int
	}{
		{
			name:     "normal width",
			bbox:     BoundingBox{X1: 10, Y1: 20, X2: 110, Y2: 30},
			expected: 100,
		},
		{
			name:     "zero width",
			bbox:     BoundingBox{X1: 50, Y1: 20, X2: 50, Y2: 30},
			expected: 0,
		},
		{
			name:     "full width",
			bbox:     BoundingBox{X1: 0, Y1: 0, X2: 1920, Y2: 1080},
			expected: 1920,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.bbox.Width()
			if result != tt.expected {
				t.Errorf("BoundingBox.Width() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestBoundingBoxHeight(t *testing.T) {
	tests := []struct {
		name     string
		bbox     BoundingBox
		expected int
	}{
		{
			name:     "normal height",
			bbox:     BoundingBox{X1: 10, Y1: 20, X2: 110, Y2: 120},
			expected: 100,
		},
		{
			name:     "zero height",
			bbox:     BoundingBox{X1: 10, Y1: 50, X2: 110, Y2: 50},
			expected: 0,
		},
		{
			name:     "full height",
			bbox:     BoundingBox{X1: 0, Y1: 0, X2: 1920, Y2: 1080},
			expected: 1080,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.bbox.Height()
			if result != tt.expected {
				t.Errorf("BoundingBox.Height() = %v, want %v", result, tt.expected)
			}
		})
	}
}
