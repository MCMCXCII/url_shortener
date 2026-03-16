package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

type compressWriter struct {
	w           http.ResponseWriter
	zw          *gzip.Writer
	compress    bool
	wroteHeader bool
	statusCode  int
}

func newCompressWriter(w http.ResponseWriter) *compressWriter {
	return &compressWriter{
		w: w,
	}
}

func (c *compressWriter) Header() http.Header {
	return c.w.Header()
}

func (c *compressWriter) Write(p []byte) (int, error) {
	// если заголовок ещё не отправлен — отправляем 200
	if !c.wroteHeader {
		c.WriteHeader(http.StatusOK)
	}

	// ❗ тело редиректов (3xx) НЕ пишем
	if c.statusCode >= 300 && c.statusCode < 400 {
		return len(p), nil
	}

	// включаем gzip только если ещё не включён
	if !c.compress {
		ct := c.w.Header().Get("Content-Type")

		if strings.HasPrefix(ct, "application/json") ||
			strings.HasPrefix(ct, "text/html") {

			c.w.Header().Set("Content-Encoding", "gzip")
			c.zw = gzip.NewWriter(c.w)
			c.compress = true
		}
	}

	if c.compress {
		return c.zw.Write(p)
	}

	return c.w.Write(p)
}

func (c *compressWriter) WriteHeader(statusCode int) {
	if c.wroteHeader {
		return
	}
	c.wroteHeader = true
	c.statusCode = statusCode

	// ❗ gzip для редиректов запрещён
	if statusCode >= 300 && statusCode < 400 {
		c.compress = false
	}

	c.w.WriteHeader(statusCode)
}

func (c *compressWriter) Close() error {
	if c.compress && c.zw != nil {
		return c.zw.Close()
	}
	return nil
}

// ---------------------------
// GZIP READER
// ---------------------------

type compressReader struct {
	r  io.ReadCloser
	zr *gzip.Reader
}

func newCompressReader(r io.ReadCloser) (*compressReader, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}

	return &compressReader{
		r:  r,
		zr: zr,
	}, nil
}

func (c compressReader) Read(p []byte) (n int, err error) {
	return c.zr.Read(p)
}

func (c *compressReader) Close() error {
	if err := c.r.Close(); err != nil {
		return err
	}
	return c.zr.Close()
}
