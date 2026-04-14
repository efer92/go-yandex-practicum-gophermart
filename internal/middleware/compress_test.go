package middleware_test

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gophermart/internal/middleware"
)

// TestCompress_GzipRequest verifies that gzip-encoded request bodies are decompressed.
func TestCompress_GzipRequest(t *testing.T) {
	var gotBody string
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		gotBody = string(body)
		w.WriteHeader(http.StatusOK)
	})

	handler := middleware.Compress(inner)

	// Compress the body.
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	gz.Write([]byte("hello world"))
	gz.Close()

	req := httptest.NewRequest(http.MethodPost, "/", &buf)
	req.Header.Set("Content-Encoding", "gzip")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "hello world", gotBody)
}

// TestCompress_GzipResponse verifies that responses are compressed when client accepts gzip.
func TestCompress_GzipResponse(t *testing.T) {
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("response body"))
	})

	handler := middleware.Compress(inner)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, "gzip", rr.Header().Get("Content-Encoding"))

	gr, err := gzip.NewReader(rr.Body)
	require.NoError(t, err)
	defer gr.Close()
	decompressed, err := io.ReadAll(gr)
	require.NoError(t, err)
	assert.Equal(t, "response body", string(decompressed))
}

// TestCompress_NoCompression verifies plain responses when client doesn't accept gzip.
func TestCompress_NoCompression(t *testing.T) {
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("plain response"))
	})

	handler := middleware.Compress(inner)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Empty(t, rr.Header().Get("Content-Encoding"))
	assert.Equal(t, "plain response", rr.Body.String())
}

// TestCompress_InvalidGzipBody verifies 400 on a malformed gzip request body.
func TestCompress_InvalidGzipBody(t *testing.T) {
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("inner handler should not be called")
	})

	handler := middleware.Compress(inner)

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString("not gzip"))
	req.Header.Set("Content-Encoding", "gzip")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}
