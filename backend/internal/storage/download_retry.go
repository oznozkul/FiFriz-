package storage

import (
	"context"
	"io"
	"time"
)

type cancelOnCloseReader struct {
	io.ReadCloser
	cancel context.CancelFunc
}

func (r *cancelOnCloseReader) Close() error {
	err := r.ReadCloser.Close()
	if r.cancel != nil {
		r.cancel()
	}
	return err
}

// DownloadWithRetry downloads object content with retry for transient MinIO failures.
func (m *Manager) DownloadWithRetry(ctx context.Context, location DocumentLocation, maxRetries int, timeout time.Duration) (io.ReadCloser, error) {
	if maxRetries <= 0 {
		maxRetries = 1
	}
	if timeout <= 0 {
		timeout = 30 * time.Second
	}

	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		attemptCtx := ctx
		cancel := func() {}
		if _, hasDeadline := ctx.Deadline(); !hasDeadline {
			attemptCtx, cancel = context.WithTimeout(ctx, timeout)
		}

		reader, err := m.Download(attemptCtx, location)
		if err != nil {
			cancel()
			lastErr = err
			time.Sleep(time.Duration(attempt+1) * 200 * time.Millisecond)
			continue
		}

		return &cancelOnCloseReader{ReadCloser: reader, cancel: cancel}, nil
	}
	return nil, lastErr
}
