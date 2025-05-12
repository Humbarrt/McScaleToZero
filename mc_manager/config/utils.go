package config

import (
	"encoding/base64"
	"fmt"
	"image"
	_ "image/png"
	"mc_manager/logging"
	"os"
	"strings"
)

func getenv(key, fallback string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return fallback
	}
	return value
}

func validateImage(base64Image *string) error {
	if base64Image == nil || len(*base64Image) == 0 {
		return fmt.Errorf("base64 image is nil")
	}

	// Clean up base64 string (remove newlines or prefix like "data:image/png;base64,")
	base64ImageStr := strings.TrimSpace(*base64Image)
	if strings.Contains(base64ImageStr, ",") {
		parts := strings.SplitN(base64ImageStr, ",", 2)
		base64ImageStr = parts[1]
	}

	// Decode Base64
	imgData, err := base64.StdEncoding.DecodeString(base64ImageStr)
	if err != nil {
		logging.Error("Error decoding Base64:", "error", err)
		return err
	}

	// Decode image
	img, _, err := image.Decode(strings.NewReader(string(imgData)))
	if err != nil {
		logging.Error("Error decoding image:", "error", err)
		return err
	}

	// Get image bounds
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	fmt.Printf("Width: %d, Height: %d\n", width, height)
	if width != 64 || height != 64 {
		return fmt.Errorf("image dimensions are not 64x64")
	}
	return nil
}
