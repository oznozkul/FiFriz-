package virusscan

import (
	"context"
	"io"
)

// ScanResult is returned after a virus scan attempt.
type ScanResult struct {
	Clean bool
	Threat string
}

// Scanner scans uploaded content. Production can swap in ClamAV.
type Scanner interface {
	Scan(ctx context.Context, reader io.Reader, fileName string) (ScanResult, error)
}

// NoOpScanner accepts all files.
type NoOpScanner struct{}

func (NoOpScanner) Scan(context.Context, io.Reader, string) (ScanResult, error) {
	return ScanResult{Clean: true}, nil
}
