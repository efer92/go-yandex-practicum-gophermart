package middleware

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

// responseWriter wraps http.ResponseWriter to capture the status code.
type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

// Logger returns middleware that logs each HTTP request using the given zap logger.
// It records the method, path, status code, duration, and user ID (if authenticated).
func Logger(log *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}

			next.ServeHTTP(rw, r)

			fields := []zap.Field{
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.Int("status", rw.status),
				zap.Duration("duration", time.Since(start)),
			}

			// Include user ID when request was authenticated.
			if uid := UserIDFromCtx(r.Context()); uid != 0 {
				fields = append(fields, zap.Int64("user_id", uid))
			}

			switch {
			case rw.status >= 500:
				log.Error("request", fields...)
			case rw.status >= 400:
				log.Warn("request", fields...)
			default:
				log.Info("request", fields...)
			}
		})
	}
}
