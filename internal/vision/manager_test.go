package vision

import (
	"testing"
)

func TestVisionQualityConstants(t *testing.T) {
	tests := []struct {
		quality VisionQuality
		str     string
	}{
		{QualityLow, "low"},
		{QualityMiddle, "middle"},
		{QualityHigh, "high"},
	}

	for _, tt := range tests {
		t.Run(string(tt.quality), func(t *testing.T) {
			if string(tt.quality) != tt.str {
				t.Errorf("VisionQuality = %v, want %v", string(tt.quality), tt.str)
			}
		})
	}
}

func TestGetQualityDescription(t *testing.T) {
	tests := []struct {
		quality  VisionQuality
		expected string
	}{
		{QualityLow, "Use minimal detail for fast processing."},
		{QualityMiddle, "Use balanced detail for processing."},
		{QualityHigh, "Use maximum detail for accurate processing."},
	}

	for _, tt := range tests {
		t.Run(string(tt.quality), func(t *testing.T) {
			result := getQualityDescription(tt.quality)
			if result != tt.expected {
				t.Errorf("getQualityDescription() = %v, want %v", result, tt.expected)
			}
		})
	}
}
