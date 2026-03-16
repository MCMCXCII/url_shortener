package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

type compressWriter struct {
	http.ResponseWriter
	zw           *gzip.Writer
	compress     bool
	wroteHeader  bool
	supportsGzip bool
}

func newCompressWriter(w http.ResponseWriter, supportsGzip bool) *compressWriter {
	return &compressWriter{
		ResponseWriter: w,
		supportsGzip:   supportsGzip,
	}
}

func (c *compressWriter) WriteHeader(statusCode int) {
	if c.wroteHeader {
		return
	}
	c.wroteHeader = true

	// редиректы не сжимаем
	if statusCode >= 300 && statusCode < 400 {
		c.compress = false
	}

	c.ResponseWriter.WriteHeader(statusCode)
}

func (c *compressWriter) Write(p []byte) (int, error) {
	if !c.wroteHeader {
		c.WriteHeader(http.StatusOK)
	}

	// редиректы — не пишем тело
	if c.compress == false && (c.ResponseWriter.Header().Get("Location") != "") {
		return len(p), nil
	}

	// решаем, включать ли gzip
	if !c.compress && c.supportsGzip {
		ct := c.ResponseWriter.Header().Get("Content-Type")

		if strings.HasPrefix(ct, "application/json") ||
			strings.HasPrefix(ct, "text/html") {

			c.ResponseWriter.Header().Set("Content-Encoding", "gzip")
			c.zw = gzip.NewWriter(c.ResponseWriter)
			c.compress = true
		}
	}

	if c.compress {
		return c.zw.Write(p)
	}

	return c.ResponseWriter.Write(p)
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
