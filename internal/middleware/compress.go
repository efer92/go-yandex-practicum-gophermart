package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

// Compress returns middleware that:
//   - Decompresses gzip-encoded request bodies (Content-Encoding: gzip).
//   - Compresses response bodies with gzip when the client accepts it (Accept-Encoding: gzip).
func Compress(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Decompress request body if gzip-encoded.
		if r.Header.Get("Content-Encoding") == "gzip" {
			gr, err := gzip.NewReader(r.Body)
			if err != nil {
				http.Error(w, "invalid gzip body", http.StatusBadRequest)
				return
			}
			defer gr.Close()
			r.Body = io.NopCloser(gr)
			r.Header.Del("Content-Encoding")
		}

		// Compress response if client accepts gzip.
		if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			gw := gzip.NewWriter(w)
			defer gw.Close()
			w.Header().Set("Content-Encoding", "gzip")
			next.ServeHTTP(&gzipResponseWriter{ResponseWriter: w, writer: gw}, r)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// gzipResponseWriter wraps http.ResponseWriter to write through a gzip.Writer.
type gzipResponseWriter struct {
	http.ResponseWriter
	writer io.Writer
}

// Write sends data through the gzip writer.
func (g *gzipResponseWriter) Write(b []byte) (int, error) {
	return g.writer.Write(b)
}
