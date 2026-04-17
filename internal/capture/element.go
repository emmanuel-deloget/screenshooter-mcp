package capture

import "image"

type BoundingBox struct {
	X1 int `json:"x1"`
	Y1 int `json:"y1"`
	X2 int `json:"x2"`
	Y2 int `json:"y2"`
}

type Element struct {
	BoundingBox BoundingBox `json:"bounding_box"`
	Description string      `json:"description"`
	Confidence  float64     `json:"confidence"`
}

func (b *BoundingBox) IsValid() bool {
	return b.X1 >= 0 && b.Y1 >= 0 && b.X2 > b.X1 && b.Y2 > b.Y1
}

func (b *BoundingBox) Width() int {
	return b.X2 - b.X1
}

func (b *BoundingBox) Height() int {
	return b.Y2 - b.Y1
}

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
