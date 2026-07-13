package stream

import (
	"io"
	"mime"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

// SendFile buffers the reader and writes a complete HTTP response.
// This avoids nginx 502 errors that can occur with chunked upstream streams.
func SendFile(c *fiber.Ctx, reader io.ReadCloser, size int64, mimeType, fileName, disposition string) error {
	if reader == nil {
		return fiber.ErrInternalServerError
	}
	defer reader.Close()

	if mimeType != "" {
		c.Set(fiber.HeaderContentType, mimeType)
	}
	c.Set(fiber.HeaderContentDisposition, mime.FormatMediaType(disposition, map[string]string{"filename": fileName}))

	data, err := io.ReadAll(reader)
	if err != nil {
		return err
	}
	if size > 0 {
		c.Set(fiber.HeaderContentLength, strconv.FormatInt(size, 10))
	} else if len(data) > 0 {
		c.Set(fiber.HeaderContentLength, strconv.Itoa(len(data)))
	}

	return c.Send(data)
}
