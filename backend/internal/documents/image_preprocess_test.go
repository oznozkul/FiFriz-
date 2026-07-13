package documents

import (
	"image"
	"image/color"
	"testing"
)

func TestParseRotateAngle(t *testing.T) {
	output := "Page number: 0\nOrientation in degrees: 90\nRotate: 90\nOrientation confidence: 5.43\n"
	if got := parseRotateAngle(output); got != 90 {
		t.Fatalf("parseRotateAngle() = %d, want 90", got)
	}
}

func TestNormalizeRotation(t *testing.T) {
	tests := map[int]int{
		0:   0,
		90:  90,
		180: 180,
		270: 270,
		360: 0,
		-90: 270,
		45:  0,
	}

	for input, want := range tests {
		if got := normalizeRotation(input); got != want {
			t.Fatalf("normalizeRotation(%d) = %d, want %d", input, got, want)
		}
	}
}

func TestAdaptiveThresholdSeparatesTextFromBackground(t *testing.T) {
	gray := image.NewGray(image.Rect(0, 0, 40, 20))
	for y := 0; y < 20; y++ {
		for x := 0; x < 40; x++ {
			value := uint8(230)
			if x >= 10 && x <= 28 && y >= 6 && y <= 13 {
				value = 30
			}
			gray.SetGray(x, y, color.Gray{Y: value})
		}
	}

	threshold := adaptiveThreshold(gray, 9, 8)
	if threshold.GrayAt(20, 10).Y != 0 {
		t.Fatalf("expected dark text pixel, got %d", threshold.GrayAt(20, 10).Y)
	}
	if threshold.GrayAt(2, 2).Y != 255 {
		t.Fatalf("expected bright background pixel, got %d", threshold.GrayAt(2, 2).Y)
	}
}

func TestUpscaleForOCRScalesSmallImages(t *testing.T) {
	small := image.NewGray(image.Rect(0, 0, 600, 400))
	upscaled := upscaleForOCR(small)
	if upscaled.Bounds().Dx() <= 600 {
		t.Fatalf("upscaleForOCR() width = %d, want > 600", upscaled.Bounds().Dx())
	}
}

func TestUpscaleForOCRKeepsLargeImages(t *testing.T) {
	large := image.NewGray(image.Rect(0, 0, 2200, 1600))
	upscaled := upscaleForOCR(large)
	if upscaled.Bounds().Dx() != 2200 {
		t.Fatalf("upscaleForOCR() width = %d, want 2200", upscaled.Bounds().Dx())
	}
}
