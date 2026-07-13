package documents

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/disintegration/imaging"
	"github.com/rwcarlsen/goexif/exif"
)

const (
	ocrMinWidth        = 1800
	ocrMaxUpscale      = 3.0
	ocrContrastPercent = 28.0
	ocrBlurSigma       = 0.7
	ocrThresholdBlock  = 31
	ocrThresholdConst  = 12
)

type preprocessConfig struct {
	debugDir string
}

type preprocessDebugger struct {
	enabled bool
	dir     string
}

func preprocessImageForOCR(sourcePath, tesseractPath string, config preprocessConfig) (string, func(), error) {
	sourceFile, err := os.Open(sourcePath)
	if err != nil {
		return "", nil, fmt.Errorf("open source image: %w", err)
	}

	img, err := imaging.Decode(sourceFile)
	closeErr := sourceFile.Close()
	if err != nil {
		return "", nil, fmt.Errorf("decode source image: %w", err)
	}
	if closeErr != nil {
		return "", nil, closeErr
	}

	debugger := newPreprocessDebugger(config.debugDir, sourcePath)
	debugger.save("01-original", img)

	img = applyEXIFOrientation(img, readEXIFOrientation(sourcePath))
	debugger.save("02-exif-oriented", img)

	rotation, rotationErr := detectTesseractRotation(tesseractPath, img, debugger.dir)
	if rotationErr != nil {
		log.Printf("ocr orientation detection skipped: %v", rotationErr)
	}
	if rotation != 0 {
		img = rotateImage(img, rotation)
	}
	debugger.save("03-oriented", img)

	var current image.Image = imaging.Grayscale(img)
	debugger.save("04-grayscale", current)

	current = upscaleForOCR(current)
	debugger.save("05-resized", current)

	current = imaging.AdjustContrast(current, ocrContrastPercent)
	debugger.save("06-contrast", current)

	current = imaging.Blur(current, ocrBlurSigma)
	debugger.save("07-denoised", current)

	threshold := adaptiveThreshold(toGrayImage(current), ocrThresholdBlock, ocrThresholdConst)
	debugger.save("08-threshold", threshold)

	processedPath, removeTemp, err := saveTemporaryPNG(threshold)
	if err != nil {
		return "", nil, err
	}

	debugger.saveFileCopy("09-ocr-input.png", processedPath)

	cleanup := func() {
		removeTemp()
	}

	return processedPath, cleanup, nil
}

func newPreprocessDebugger(debugDir, sourcePath string) *preprocessDebugger {
	debugDir = strings.TrimSpace(debugDir)
	if debugDir == "" {
		return &preprocessDebugger{}
	}

	baseName := strings.TrimSuffix(filepath.Base(sourcePath), filepath.Ext(sourcePath))
	sessionDir := filepath.Join(debugDir, fmt.Sprintf("%s-%d", sanitizeDebugName(baseName), time.Now().UnixNano()))

	if err := os.MkdirAll(sessionDir, 0o755); err != nil {
		log.Printf("ocr debug directory could not be created: %v", err)
		return &preprocessDebugger{}
	}

	return &preprocessDebugger{
		enabled: true,
		dir:     sessionDir,
	}
}

func sanitizeDebugName(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "image"
	}

	replacer := strings.NewReplacer(
		"\\", "_",
		"/", "_",
		":", "_",
		"*", "_",
		"?", "_",
		"\"", "_",
		"<", "_",
		">", "_",
		"|", "_",
		" ", "_",
	)
	return replacer.Replace(value)
}

func (d *preprocessDebugger) save(stage string, img image.Image) {
	if !d.enabled || img == nil {
		return
	}

	targetPath := filepath.Join(d.dir, stage+".png")
	if err := imaging.Save(img, targetPath); err != nil {
		log.Printf("ocr debug save failed for %s: %v", stage, err)
	}
}

func (d *preprocessDebugger) saveFileCopy(stage string, sourcePath string) {
	if !d.enabled {
		return
	}

	data, err := os.ReadFile(sourcePath)
	if err != nil {
		log.Printf("ocr debug copy read failed for %s: %v", stage, err)
		return
	}

	targetPath := filepath.Join(d.dir, stage)
	if err := os.WriteFile(targetPath, data, 0o644); err != nil {
		log.Printf("ocr debug copy write failed for %s: %v", stage, err)
	}
}

func readEXIFOrientation(sourcePath string) int {
	file, err := os.Open(sourcePath)
	if err != nil {
		return 1
	}
	defer file.Close()

	metadata, err := exif.Decode(file)
	if err != nil {
		return 1
	}

	tag, err := metadata.Get(exif.Orientation)
	if err != nil {
		return 1
	}

	orientation, err := tag.Int(0)
	if err != nil || orientation < 1 || orientation > 8 {
		return 1
	}

	return orientation
}

func applyEXIFOrientation(img image.Image, orientation int) image.Image {
	switch orientation {
	case 2:
		return imaging.FlipH(img)
	case 3:
		return imaging.Rotate180(img)
	case 4:
		return imaging.FlipV(img)
	case 5:
		return imaging.Rotate270(imaging.FlipH(img))
	case 6:
		return imaging.Rotate270(img)
	case 7:
		return imaging.Rotate90(imaging.FlipH(img))
	case 8:
		return imaging.Rotate90(img)
	default:
		return img
	}
}

func detectTesseractRotation(tesseractPath string, img image.Image, debugDir string) (int, error) {
	osdPath, cleanup, err := saveTemporaryPNG(img)
	if err != nil {
		return 0, err
	}
	defer cleanup()

	rotation, detectErr := detectRotationFromOSD(tesseractPath, osdPath)
	if detectErr != nil {
		return 0, detectErr
	}

	if debugDir != "" {
		log.Printf("ocr orientation detected: rotate=%d debug_dir=%s", rotation, debugDir)
	}

	return rotation, nil
}

func detectRotationFromOSD(tesseractPath, imagePath string) (int, error) {
	output, err := runTesseractCommand(tesseractPath, imagePath, "stdout", "-l", "osd", "--psm", "0")
	if err != nil {
		return 0, err
	}

	rotation := parseRotateAngle(string(output))
	if rotation == 0 {
		return 0, fmt.Errorf("orientation not detected")
	}

	return rotation, nil
}

func parseRotateAngle(output string) int {
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(strings.ToLower(line), "rotate:") {
			continue
		}

		value := strings.TrimSpace(line[strings.Index(line, ":")+1:])
		angle, err := strconv.Atoi(value)
		if err != nil {
			continue
		}

		return normalizeRotation(angle)
	}

	return 0
}

func normalizeRotation(angle int) int {
	angle = angle % 360
	if angle < 0 {
		angle += 360
	}

	switch angle {
	case 90, 180, 270:
		return angle
	default:
		return 0
	}
}

func rotateImage(img image.Image, degrees int) image.Image {
	switch degrees {
	case 90:
		return imaging.Rotate90(img)
	case 180:
		return imaging.Rotate180(img)
	case 270:
		return imaging.Rotate270(img)
	default:
		return img
	}
}

func upscaleForOCR(img image.Image) image.Image {
	bounds := img.Bounds()
	width := bounds.Dx()
	if width >= ocrMinWidth {
		return img
	}

	scale := float64(ocrMinWidth) / float64(width)
	if scale > ocrMaxUpscale {
		scale = ocrMaxUpscale
	}

	targetWidth := int(math.Round(float64(width) * scale))
	if targetWidth <= width {
		return img
	}

	return imaging.Resize(img, targetWidth, 0, imaging.Lanczos)
}

func toGrayImage(img image.Image) *image.Gray {
	bounds := img.Bounds()
	gray := image.NewGray(bounds)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			gray.Set(x, y, color.GrayModel.Convert(img.At(x, y)).(color.Gray))
		}
	}
	return gray
}

func adaptiveThreshold(gray *image.Gray, blockSize, constant int) *image.Gray {
	if blockSize%2 == 0 {
		blockSize++
	}
	if blockSize < 3 {
		blockSize = 3
	}

	bounds := gray.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	output := image.NewGray(bounds)

	integral := make([]uint64, (width+1)*(height+1))
	for y := 0; y < height; y++ {
		rowSum := uint64(0)
		for x := 0; x < width; x++ {
			rowSum += uint64(gray.GrayAt(bounds.Min.X+x, bounds.Min.Y+y).Y)
			top := integral[y*(width+1)+x]
			outputIndex := (y+1)*(width+1) + (x + 1)
			integral[outputIndex] = rowSum + top
		}
	}

	halfBlock := blockSize / 2
	for y := 0; y < height; y++ {
		y1 := maxInt(0, y-halfBlock)
		y2 := minInt(height-1, y+halfBlock)
		for x := 0; x < width; x++ {
			x1 := maxInt(0, x-halfBlock)
			x2 := minInt(width-1, x+halfBlock)

			count := (x2 - x1 + 1) * (y2 - y1 + 1)
			sum := boxSum(integral, width, x1, y1, x2, y2)
			threshold := int(sum/uint64(count)) - constant

			pixel := int(gray.GrayAt(bounds.Min.X+x, bounds.Min.Y+y).Y)
			if pixel < threshold {
				output.SetGray(bounds.Min.X+x, bounds.Min.Y+y, color.Gray{Y: 0})
			} else {
				output.SetGray(bounds.Min.X+x, bounds.Min.Y+y, color.Gray{Y: 255})
			}
		}
	}

	return output
}

func boxSum(integral []uint64, width, x1, y1, x2, y2 int) uint64 {
	a := integral[y1*(width+1)+x1]
	b := integral[y1*(width+1)+(x2+1)]
	c := integral[(y2+1)*(width+1)+x1]
	d := integral[(y2+1)*(width+1)+(x2+1)]
	return d + a - b - c
}

func saveTemporaryPNG(img image.Image) (string, func(), error) {
	tempFile, err := os.CreateTemp("", "fifriz-ocr-*.png")
	if err != nil {
		return "", nil, err
	}

	tempPath := tempFile.Name()
	if err := png.Encode(tempFile, img); err != nil {
		_ = tempFile.Close()
		_ = os.Remove(tempPath)
		return "", nil, err
	}
	if err := tempFile.Close(); err != nil {
		_ = os.Remove(tempPath)
		return "", nil, err
	}

	cleanup := func() {
		_ = os.Remove(tempPath)
	}

	return tempPath, cleanup, nil
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
