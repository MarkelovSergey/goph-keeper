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
