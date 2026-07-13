package documents

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

var (
	ErrTesseractNotAvailable = errors.New("tesseract ocr is not available")
	ErrOCRFailed             = errors.New("ocr failed")
)

func extractImageText(filePath string) (string, error) {
	tesseractPath, err := resolveTesseractCommand()
	if err != nil {
		return "", err
	}

	ocrInputPath := filePath
	cleanup := func() {}

	processedPath, processedCleanup, preprocessErr := preprocessImageForOCR(filePath, tesseractPath, preprocessConfig{
		debugDir: strings.TrimSpace(os.Getenv("OCR_DEBUG_DIR")),
	})
	if preprocessErr != nil {
		log.Printf("ocr preprocessing failed, using original image: path=%s error=%v", filePath, preprocessErr)
	} else {
		ocrInputPath = processedPath
		cleanup = processedCleanup
	}
	defer cleanup()

	return runTesseractOCR(tesseractPath, ocrInputPath)
}

func runTesseractOCR(tesseractPath, imagePath string) (string, error) {
	language := strings.TrimSpace(os.Getenv("TESSERACT_LANG"))
	if language == "" {
		language = "tur+eng"
	}

	output, cmdErr := runTesseractCommand(
		tesseractPath,
		imagePath,
		"stdout",
		"-l", language,
		"--psm", "3",
		"--oem", "1",
	)
	if cmdErr != nil {
		if errors.Is(cmdErr, exec.ErrNotFound) {
			return "", ErrTesseractNotAvailable
		}

		if exitErr, ok := cmdErr.(*exec.ExitError); ok {
			stderr := strings.TrimSpace(string(exitErr.Stderr))
			if isTesseractMissingError(stderr, cmdErr) {
				return "", ErrTesseractNotAvailable
			}
			if stderr != "" {
				return "", fmt.Errorf("%w: %s", ErrOCRFailed, stderr)
			}
		}

		return "", fmt.Errorf("%w: %v", ErrOCRFailed, cmdErr)
	}

	return normalizeExtractedText(output), nil
}

func runTesseractCommand(tesseractPath, imagePath string, args ...string) ([]byte, error) {
	commandArgs := append([]string{imagePath}, args...)
	cmd := exec.Command(tesseractPath, commandArgs...)
	return cmd.Output()
}

func resolveTesseractCommand() (string, error) {
	if configured := strings.TrimSpace(os.Getenv("TESSERACT_CMD")); configured != "" {
		if path, err := resolveExecutablePath(configured); err == nil {
			return path, nil
		}
	}

	if path, err := exec.LookPath("tesseract"); err == nil {
		return path, nil
	}

	for _, candidate := range defaultTesseractPaths() {
		if path, err := resolveExecutablePath(candidate); err == nil {
			return path, nil
		}
	}

	return "", ErrTesseractNotAvailable
}

func defaultTesseractPaths() []string {
	if runtime.GOOS == "windows" {
		return []string{
			`C:\Program Files\Tesseract-OCR\tesseract.exe`,
			`C:\Program Files (x86)\Tesseract-OCR\tesseract.exe`,
		}
	}

	return []string{
		"/usr/bin/tesseract",
		"/usr/local/bin/tesseract",
		"/opt/homebrew/bin/tesseract",
	}
}

func resolveExecutablePath(path string) (string, error) {
	if strings.Contains(path, string(os.PathSeparator)) {
		info, err := os.Stat(path)
		if err != nil {
			return "", err
		}
		if info.IsDir() {
			return "", fmt.Errorf("path is a directory: %s", path)
		}
		return path, nil
	}

	return exec.LookPath(path)
}

func isTesseractMissingError(stderr string, err error) bool {
	lowered := strings.ToLower(stderr + " " + err.Error())
	missingMarkers := []string{
		"not recognized as an internal or external command",
		"executable file not found",
		"no such file or directory",
		"failed to spawn",
		"cannot find the file specified",
	}

	for _, marker := range missingMarkers {
		if strings.Contains(lowered, marker) {
			return true
		}
	}

	return false
}

func isImageExtension(extension string) bool {
	switch normalizeStoredExtension(extension) {
	case "png", "jpg", "jpeg":
		return true
	default:
		return false
	}
}
