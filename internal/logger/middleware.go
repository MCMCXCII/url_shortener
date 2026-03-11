package logger

import (
	"net/http"
	"time"

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
		Log.Debug("request",
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

		Log.Debug("response",
			zap.Int("status code", responseData.code),
			zap.Int("size", responseData.size))
	})
}
