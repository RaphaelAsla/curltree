package handlers

import (
	"net/http"
	"time"

	"curltree/pkg/utils"
)

type LoggingMiddleware struct {
	logger *utils.Logger
}

func NewLoggingMiddleware(logger *utils.Logger) *LoggingMiddleware {
	return &LoggingMiddleware{
		logger: logger.WithContext("http"),
	}
}

func (lm *LoggingMiddleware) Middleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		
		recorder := &responseRecorder{
			ResponseWriter: w,
			statusCode:    http.StatusOK,
		}
		
		next.ServeHTTP(recorder, r)
		
		duration := time.Since(start)
		
		lm.logger.LogRequest(
			r.Method,
			r.URL.Path,
			recorder.statusCode,
			duration.String(),
		)
		
		if recorder.statusCode >= 400 {
			lm.logger.Error("HTTP error response",
				"method", r.Method,
				"path", r.URL.Path,
				"status_code", recorder.statusCode,
				"user_agent", r.UserAgent(),
				"remote_addr", getClientIP(r),
			)
		}
	}
}

type responseRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (rr *responseRecorder) WriteHeader(statusCode int) {
	rr.statusCode = statusCode
	rr.ResponseWriter.WriteHeader(statusCode)
}