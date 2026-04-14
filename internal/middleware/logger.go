package middleware

import (
	"context"
	"net/http"
	"time"

	"go.uber.org/zap"
)

type logFieldsKey struct{}

// responseWriter wraps http.ResponseWriter to capture the status code.
type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

// AddLogFields appends extra zap fields that the Logger middleware will include
// in the log line for this request. The Logger must have run first (it seeds the
// pointer in the context); calling this outside of a Logger-wrapped handler is a
// no-op.
func AddLogFields(r *http.Request, fields ...zap.Field) {
	ptr, _ := r.Context().Value(logFieldsKey{}).(*[]zap.Field)
	if ptr == nil {
		return
	}
	*ptr = append(*ptr, fields...)
}

// Logger returns middleware that logs each HTTP request using the given zap logger.
// It records the method, path, status code, duration, user ID, and any extra
// fields added by handlers via AddLogFields.
func Logger(log *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}

			// Seed a mutable slice pointer so handlers can append fields via AddLogFields.
			extra := &[]zap.Field{}
			r = r.WithContext(context.WithValue(r.Context(), logFieldsKey{}, extra))

			next.ServeHTTP(rw, r)

			fields := []zap.Field{
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.Int("status", rw.status),
				zap.Duration("duration", time.Since(start)),
			}

			if uid := UserIDFromCtx(r.Context()); uid != "" {
				fields = append(fields, zap.String("user_id", uid))
			}

			fields = append(fields, *extra...)

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
