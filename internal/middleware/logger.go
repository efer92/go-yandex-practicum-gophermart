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

// AddLogFields attaches extra zap fields to the request context.
// Handlers call this to enrich the log line written by the Logger middleware.
func AddLogFields(r *http.Request, fields ...zap.Field) *http.Request {
	existing, _ := r.Context().Value(logFieldsKey{}).([]zap.Field)
	updated := make([]zap.Field, len(existing)+len(fields))
	copy(updated, existing)
	copy(updated[len(existing):], fields)
	return r.WithContext(context.WithValue(r.Context(), logFieldsKey{}, updated))
}

// logFieldsFromCtx retrieves extra log fields stored by AddLogFields.
func logFieldsFromCtx(ctx context.Context) []zap.Field {
	fields, _ := ctx.Value(logFieldsKey{}).([]zap.Field)
	return fields
}

// Logger returns middleware that logs each HTTP request using the given zap logger.
// It records the method, path, status code, duration, user ID, and any extra
// fields added by handlers via AddLogFields.
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

			if uid := UserIDFromCtx(r.Context()); uid != 0 {
				fields = append(fields, zap.Int64("user_id", uid))
			}

			fields = append(fields, logFieldsFromCtx(r.Context())...)

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
