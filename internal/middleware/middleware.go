package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/MCMCXCII/url_shortener/internal/logger"
	"go.uber.org/zap"
)

type responseData struct {
	code int
	size int
}

type loggingResponseWritter struct {
	http.ResponseWriter
	responseData *responseData
}

func (r *loggingResponseWritter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size
	return size, err
}

func (r *loggingResponseWritter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.code = statusCode
}

func RequestLogger(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		h.ServeHTTP(w, r)
		duration := time.Since(start)
		logger.Log.Debug("request",
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
			zap.Duration("duration", duration))
	})
}

func ResponseLogger(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		responseData := &responseData{
			code: 0,
			size: 0,
		}
		lw := loggingResponseWritter{
			ResponseWriter: w,
			responseData:   responseData,
		}
		h.ServeHTTP(&lw, r)

		logger.Log.Debug("response",
			zap.Int("status code", responseData.code),
			zap.Int("size", responseData.size))
	})
}

func GzipMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// 1. Распаковка входящего gzip
		if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
			cr, err := newCompressReader(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			r.Body = cr
			defer cr.Close()
		}

		// 2. Проверяем, поддерживает ли клиент gzip
		supportsGzip := strings.Contains(r.Header.Get("Accept-Encoding"), "gzip")

		// 3. Если не поддерживает — просто вызываем хендлер
		if !supportsGzip {
			h.ServeHTTP(w, r)
			return
		}

		// 4. Оборачиваем writer, но gzip включится ТОЛЬКО после WriteHeader
		cw := newCompressWriter(w, true)
		defer cw.Close()

		h.ServeHTTP(cw, r)
	})
}
