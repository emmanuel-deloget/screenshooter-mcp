// Package element provides types for element/region handling in screen capture.
//
// This package contains types for representing rectangular regions and detected elements.
// It is designed to support OCR-based detection and coordinate-based operations.
//
// Types provided:
//   - BoundingBox: Represents a rectangular region with coordinates
//   - Element: Represents a detected element with position and description
//
// Helper functions are provided for validation, dimension calculation,
// and image cropping.
package capture

import "image"

// BoundingBox represents a rectangular region.
//
// BoundingBox defines a rectangle using the coordinates of two corners:
// (X1, Y1) is the top-left corner, and (X2, Y2) is the bottom-right corner.
// The coordinate system follows the screen coordinate space where (0, 0)
// is the top-left corner of the screen.
//
// The JSON tags enable marshaling/unmarshaling for API responses.
type BoundingBox struct {
	// X1 is the X coordinate of the top-left corner.
	X1 int `json:"x1"`
	// Y1 is the Y coordinate of the top-left corner.
	Y1 int `json:"y1"`
	// X2 is the X coordinate of the bottom-right corner.
	// Must be greater than X1 for a valid box.
	X2 int `json:"x2"`
	// Y2 is the Y coordinate of the bottom-right corner.
	// Must be greater than Y1 for a valid box.
	Y2 int `json:"y2"`
}

// Element represents a detected UI element.
//
// Element represents a region of interest detected in the screen capture,
// such as text recognized by OCR, a button, or other UI component.
//
// The BoundingBox defines the element's position on screen.
// The Description contains the recognized text or other identification.
// Confidence is a value between 0 and 1 indicating how confident the
// detection system is in this result.
type Element struct {
	// BoundingBox defines the element's position on screen.
	BoundingBox BoundingBox `json:"bounding_box"`

	// Description contains the element's identified text or label.
	// This may be OCR-recognized text, a button label, etc.
	Description string `json:"description"`

	// Confidence indicates the detection confidence.
	// Values range from 0 (no confidence) to 1 (certain).
	Confidence float64 `json:"confidence"`
}

// IsValid checks if the bounding box is valid.
//
// A bounding box is valid if:
//   - Both X1 and Y1 are non-negative (not above the screen origin)
//   - X2 > X1 (positive width)
//   - Y2 > Y1 (positive height)
//
// Returns true if valid, false otherwise.
func (b *BoundingBox) IsValid() bool {
	return b.X1 >= 0 && b.Y1 >= 0 && b.X2 > b.X1 && b.Y2 > b.Y1
}

// Width returns the width of the bounding box.
//
// Returns X2 - X1, which is the horizontal dimension in pixels.
func (b *BoundingBox) Width() int {
	return b.X2 - b.X1
}

// Height returns the height of the bounding box.
//
// Returns Y2 - Y1, which is the vertical dimension in pixels.
func (b *BoundingBox) Height() int {
	return b.Y2 - b.Y1
}

// CropToBoundingBox crops an image to the specified bounding box.
//
// This function extracts a rectangular sub-image from the source image.
// The bbox coordinates are interpreted as offsets from the image's
// minimum bounds (Min.X, Min.Y).
//
// If the bounding box extends beyond the image bounds, it is clipped.
// If the bounding box starts before the image bounds, the original
// image is returned unchanged.
//
// The source image must implement the image.SubImageer interface
// to support cropping. Standard image types (image.RGBA, image.NRGBA)
// support this.
func CropToBoundingBox(img image.Image, bbox *BoundingBox) image.Image {
	bounds := img.Bounds()
	if bbox.X1 < bounds.Min.X || bbox.Y1 < bounds.Min.Y {
		return img
	}
	x1 := bounds.Min.X + bbox.X1
	y1 := bounds.Min.Y + bbox.Y1
	x2 := bounds.Min.X + bbox.X2
	y2 := bounds.Min.Y + bbox.Y2
	if x2 > bounds.Max.X {
		x2 = bounds.Max.X
	}
	if y2 > bounds.Max.Y {
		y2 = bounds.Max.Y
	}
	return img.(interface {
		SubImage(r image.Rectangle) image.Image
	}).SubImage(image.Rect(x1, y1, x2, y2))
}
