package compression

import (
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type Compressor interface {
	Compress(sourcePath, destinationPath string) error
	Decompress(sourcePath, destinationPath string) error
	ShouldCompress(extension string) bool
}

type GzipCompressor struct{}

func NewGzipCompressor() Compressor {
	return &GzipCompressor{}
}

var compressibleExtensions = map[string]struct{}{
	"txt": {}, "csv": {}, "json": {}, "xml": {}, "html": {}, "htm": {}, "eml": {},
}

var nonCompressibleExtensions = map[string]struct{}{
	"pdf": {}, "docx": {}, "xlsx": {}, "png": {}, "jpg": {}, "jpeg": {},
}

func (c *GzipCompressor) ShouldCompress(extension string) bool {
	ext := strings.ToLower(strings.TrimPrefix(strings.TrimSpace(extension), "."))
	if _, skip := nonCompressibleExtensions[ext]; skip {
		return false
	}
	_, ok := compressibleExtensions[ext]
	return ok
}

func (c *GzipCompressor) Compress(sourcePath, destinationPath string) error {
	source, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	defer source.Close()

	if err := os.MkdirAll(filepath.Dir(destinationPath), 0o755); err != nil {
		return err
	}

	destination, err := os.Create(destinationPath)
	if err != nil {
		return err
	}
	defer destination.Close()

	writer := gzip.NewWriter(destination)
	if _, err := io.Copy(writer, source); err != nil {
		_ = writer.Close()
		return err
	}
	return writer.Close()
}

func (c *GzipCompressor) Decompress(sourcePath, destinationPath string) error {
	source, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	defer source.Close()

	reader, err := gzip.NewReader(source)
	if err != nil {
		return err
	}
	defer reader.Close()

	if err := os.MkdirAll(filepath.Dir(destinationPath), 0o755); err != nil {
		return err
	}

	destination, err := os.Create(destinationPath)
	if err != nil {
		return err
	}
	defer destination.Close()

	_, err = io.Copy(destination, reader)
	return err
}
