package stream

import "io"

type autoCloseReader struct {
	reader io.Reader
	closer io.Closer
	closed bool
}

func AutoClose(r io.Reader, closer io.Closer) io.Reader {
	if r == nil {
		return nil
	}
	if closer == nil {
		if c, ok := r.(io.Closer); ok {
			closer = c
		}
	}
	return &autoCloseReader{reader: r, closer: closer}
}

func (a *autoCloseReader) Read(p []byte) (int, error) {
	if a.closed {
		return 0, io.EOF
	}

	n, err := a.reader.Read(p)
	if err != nil {
		a.close()
	}
	return n, err
}

func (a *autoCloseReader) close() {
	if a.closed {
		return
	}
	a.closed = true
	if a.closer != nil {
		_ = a.closer.Close()
	}
}
