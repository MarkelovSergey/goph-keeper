package middleware_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/MarkelovSergey/goph-keeper/internal/server/middleware"
)

func TestMaxBodySize_WithinLimit(t *testing.T) {
	body := strings.NewReader(`{"key":"value"}`)

	var readBody []byte
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var err error
		readBody, err = io.ReadAll(r.Body)
		require.NoError(t, err)
		w.WriteHeader(http.StatusOK)
	})

	handler := middleware.MaxBodySize(next)
	req := httptest.NewRequest(http.MethodPost, "/", body)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, `{"key":"value"}`, string(readBody))
}

func TestMaxBodySize_ExceedsLimit(t *testing.T) {
	// 1 МБ + 1 байт
	oversized := strings.NewReader(strings.Repeat("x", (1<<20)+1))

	var readErr error
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, readErr = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
	})

	handler := middleware.MaxBodySize(next)
	req := httptest.NewRequest(http.MethodPost, "/", oversized)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	var maxBytesErr *http.MaxBytesError
	require.ErrorAs(t, readErr, &maxBytesErr, "ожидается *http.MaxBytesError при превышении лимита")
}

func TestMaxBodySize_EmptyBody(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		assert.Empty(t, data)
		w.WriteHeader(http.StatusOK)
	})

	handler := middleware.MaxBodySize(next)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestMaxBodySize_ExactLimit(t *testing.T) {
	exactly1MB := strings.NewReader(strings.Repeat("x", 1<<20))

	var n int
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		n = len(data)
		w.WriteHeader(http.StatusOK)
	})

	handler := middleware.MaxBodySize(next)
	req := httptest.NewRequest(http.MethodPost, "/", exactly1MB)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, 1<<20, n)
}
