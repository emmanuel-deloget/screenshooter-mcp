package vision

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
)

func encodeImage(img image.Image, quality VisionQuality) ([]byte, string, error) {
	var format string
	var buf bytes.Buffer

	switch quality {
	case QualityLow:
		format = "jpeg"
		opts := jpeg.Options{Quality: 50}
		if err := jpeg.Encode(&buf, img, &opts); err != nil {
			return nil, "", fmt.Errorf("failed to encode as jpeg: %w", err)
		}
	case QualityMiddle:
		format = "png"
		if err := png.Encode(&buf, img); err != nil {
			return nil, "", fmt.Errorf("failed to encode as png: %w", err)
		}
	case QualityHigh:
		format = "png"
		if err := png.Encode(&buf, img); err != nil {
			return nil, "", fmt.Errorf("failed to encode as png: %w", err)
		}
	default:
		format = "png"
		if err := png.Encode(&buf, img); err != nil {
			return nil, "", fmt.Errorf("failed to encode as png: %w", err)
		}
	}

	encoded := base64.StdEncoding.EncodeToString(buf.Bytes())
	return []byte(encoded), format, nil
}
