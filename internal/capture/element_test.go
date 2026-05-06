package capture

import (
	"image"
	"image/color"
	"testing"
)

func TestBoundingBoxIsValid(t *testing.T) {
	tests := []struct {
		name string
		bbox BoundingBox
		want bool
	}{
		{
			name: "valid box",
			bbox: BoundingBox{X1: 10, Y1: 20, X2: 110, Y2: 120},
			want: true,
		},
		{
			name: "negative X1",
			bbox: BoundingBox{X1: -1, Y1: 20, X2: 110, Y2: 120},
			want: false,
		},
		{
			name: "negative Y1",
			bbox: BoundingBox{X1: 10, Y1: -1, X2: 110, Y2: 120},
			want: false,
		},
		{
			name: "X2 not greater than X1",
			bbox: BoundingBox{X1: 10, Y1: 20, X2: 10, Y2: 120},
			want: false,
		},
		{
			name: "Y2 not greater than Y1",
			bbox: BoundingBox{X1: 10, Y1: 20, X2: 110, Y2: 20},
			want: false,
		},
		{
			name: "zero width",
			bbox: BoundingBox{X1: 10, Y1: 20, X2: 10, Y2: 40},
			want: false,
		},
		{
			name: "zero height",
			bbox: BoundingBox{X1: 10, Y1: 20, X2: 110, Y2: 20},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.bbox.IsValid(); got != tt.want {
				t.Errorf("BoundingBox.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBoundingBoxWidth(t *testing.T) {
	tests := []struct {
		name string
		bbox BoundingBox
		want int
	}{
		{
			name: "normal box",
			bbox: BoundingBox{X1: 10, Y1: 20, X2: 110, Y2: 120},
			want: 100,
		},
		{
			name: "zero width box",
			bbox: BoundingBox{X1: 10, Y1: 20, X2: 10, Y2: 120},
			want: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.bbox.Width(); got != tt.want {
				t.Errorf("BoundingBox.Width() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBoundingBoxHeight(t *testing.T) {
	tests := []struct {
		name string
		bbox BoundingBox
		want int
	}{
		{
			name: "normal box",
			bbox: BoundingBox{X1: 10, Y1: 20, X2: 110, Y2: 120},
			want: 100,
		},
		{
			name: "zero height box",
			bbox: BoundingBox{X1: 10, Y1: 20, X2: 110, Y2: 20},
			want: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.bbox.Height(); got != tt.want {
				t.Errorf("BoundingBox.Height() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCropToBoundingBox(t *testing.T) {
	// Create a 100x100 image
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	for y := 0; y < 100; y++ {
		for x := 0; x < 100; x++ {
			img.Set(x, y, color.RGBA{R: uint8(x), G: uint8(y), B: 0, A: 255})
		}
	}

	tests := []struct {
		name    string
		bbox    *BoundingBox
		wantW   int
		wantH   int
		wantErr bool
	}{
		{
			name:  "crop normal region",
			bbox:  &BoundingBox{X1: 10, Y1: 10, X2: 50, Y2: 50},
			wantW: 40,
			wantH: 40,
		},
		{
			name:  "crop region at origin",
			bbox:  &BoundingBox{X1: 0, Y1: 0, X2: 50, Y2: 50},
			wantW: 50,
			wantH: 50,
		},
		{
			name:  "bbox starts before image bounds",
			bbox:  &BoundingBox{X1: -10, Y1: -10, X2: 50, Y2: 50},
			wantW: 100,
			wantH: 100,
		},
		{
			name:  "crop extends beyond image bounds",
			bbox:  &BoundingBox{X1: 50, Y1: 50, X2: 150, Y2: 150},
			wantW: 50,
			wantH: 50,
		},
		{
			name:  "crop completely outside",
			bbox:  &BoundingBox{X1: 200, Y1: 200, X2: 300, Y2: 300},
			wantW: 0,
			wantH: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CropToBoundingBox(img, tt.bbox)
			bounds := result.Bounds()
			if bounds.Dx() != tt.wantW {
				t.Errorf("Cropped width = %v, want %v", bounds.Dx(), tt.wantW)
			}
			if bounds.Dy() != tt.wantH {
				t.Errorf("Cropped height = %v, want %v", bounds.Dy(), tt.wantH)
			}
		})
	}
}

func TestElement(t *testing.T) {
	elem := Element{
		BoundingBox: BoundingBox{
			X1: 10, Y1: 20, X2: 110, Y2: 120,
		},
		Description: "Submit button",
		Confidence:  0.95,
	}

	if elem.BoundingBox.X1 != 10 {
		t.Errorf("Element.BoundingBox.X1 = %v, want 10", elem.BoundingBox.X1)
	}
	if elem.Description != "Submit button" {
		t.Errorf("Element.Description = %v, want 'Submit button'", elem.Description)
	}
	if elem.Confidence != 0.95 {
		t.Errorf("Element.Confidence = %v, want 0.95", elem.Confidence)
	}
}
