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
		ow := w
		acceptEncoding := r.Header.Get("Accept-Encoding")
		supportsGzip := strings.Contains(acceptEncoding, "gzip")
		if !supportsGzip {
			h.ServeHTTP(w, r)
			return
		}
		// оборачиваем оригинальный http.ResponseWriter новым с поддержкой сжатия
		cw := newCompressWriter(w)
		// меняем оригинальный http.ResponseWriter на новый
		ow = cw
		// не забываем отправить клиенту все сжатые данные после завершения middleware
		defer cw.Close()

		// проверяем, что клиент отправил серверу сжатые данные в формате gzip
		contentEncoding := r.Header.Get("Content-Encoding")
		sendsGzip := strings.Contains(contentEncoding, "gzip")
		if sendsGzip {
			// оборачиваем тело запроса в io.Reader с поддержкой декомпрессии
			cr, err := newCompressReader(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			// меняем тело запроса на новое
			r.Body = cr
			defer cr.Close()
		}

		// передаём управление хендлеру
		h.ServeHTTP(ow, r)
	})
}
