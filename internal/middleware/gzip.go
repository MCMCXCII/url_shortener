package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

type compressWriter struct {
	rw          http.ResponseWriter
	zw          *gzip.Writer
	compressed  bool
	wroteHeader bool
}

func newCompressWriter(w http.ResponseWriter) *compressWriter {
	return &compressWriter{rw: w}
}

func (c *compressWriter) Header() http.Header {
	return c.rw.Header()
}

func (c *compressWriter) WriteHeader(statusCode int) {
	if c.wroteHeader {
		return
	}
	c.wroteHeader = true

	// редиректы не сжимаем
	if statusCode >= 300 && statusCode < 400 {
		c.compressed = false
	}

	c.rw.WriteHeader(statusCode)
}

func (c *compressWriter) Write(p []byte) (int, error) {
	if !c.wroteHeader {
		c.WriteHeader(http.StatusOK)
	}

	// редирект — не пишем тело
	if c.rw.Header().Get("Location") != "" {
		return len(p), nil
	}

	// включаем gzip только если Content-Type уже известен
	if !c.compressed {
		ct := c.rw.Header().Get("Content-Type")

		if strings.HasPrefix(ct, "application/json") ||
			strings.HasPrefix(ct, "text/html") {

			c.rw.Header().Set("Content-Encoding", "gzip")
			c.zw = gzip.NewWriter(c.rw)
			c.compressed = true
		}
	}

	if c.compressed {
		return c.zw.Write(p)
	}

	return c.rw.Write(p)
}

func (c *compressWriter) Close() error {
	if c.compressed && c.zw != nil {
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
