package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

type compressWriter struct {
	w          http.ResponseWriter
	zw         *gzip.Writer
	statusCode int
}

var compressibleTypes = []string{
	"application/json",
	"text/html",
}

func isCompressible(contentType string) bool {
	for _, t := range compressibleTypes {
		if strings.Contains(contentType, t) {
			return true
		}
	}
	return false
}

func newCompressWriter(w http.ResponseWriter) *compressWriter {
	return &compressWriter{
		w: w,
	}
}

func (c *compressWriter) Header() http.Header {
	return c.w.Header()
}

func (c *compressWriter) WriteHeader(code int) {
	c.statusCode = code
	c.w.WriteHeader(code)
}
func (c *compressWriter) Write(p []byte) (int, error) {

	// не сжимаем редиректы и ошибки
	if c.statusCode >= 300 {
		return c.w.Write(p)
	}

	contentType := c.w.Header().Get("Content-Type")

	if c.zw == nil && isCompressible(contentType) {
		c.zw = gzip.NewWriter(c.w)
		c.w.Header().Set("Content-Encoding", "gzip")
	}

	if c.zw != nil {
		return c.zw.Write(p)
	}

	return c.w.Write(p)
}

func (c *compressWriter) Close() error {
	if c.zw != nil {
		return c.zw.Close()
	}
	return nil
}

type compressReader struct {
	r  io.ReadCloser
	rw *gzip.Reader
}

func newCompressReader(r io.ReadCloser) (*compressReader, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}
	return &compressReader{
		r:  r,
		rw: zr,
	}, nil
}

func (c compressReader) Read(p []byte) (int, error) {
	return c.rw.Read(p)
}

func (c *compressReader) Close() error {
	if err := c.r.Close(); err != nil {
		return err
	}
	return c.rw.Close()
}
