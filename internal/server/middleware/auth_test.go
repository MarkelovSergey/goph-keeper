package middleware_test

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/MarkelovSergey/goph-keeper/internal/server/middleware"
	"github.com/MarkelovSergey/goph-keeper/internal/server/repository"
	"github.com/MarkelovSergey/goph-keeper/internal/server/service"

	"github.com/MarkelovSergey/goph-keeper/internal/model"
)

type stubRepo struct {
	users map[string]*model.User
}

func (r *stubRepo) Create(_ context.Context, user *model.User) error {
	r.users[user.Login] = user
	return nil
}

func (r *stubRepo) GetByLogin(_ context.Context, login string) (*model.User, error) {
	if u, ok := r.users[login]; ok {
		return u, nil
	}
	return nil, repository.ErrNotFound
}

func (r *stubRepo) GetByID(_ context.Context, id uuid.UUID) (*model.User, error) {
	for _, u := range r.users {
		if u.ID == id {
			return u, nil
		}
	}
	return nil, repository.ErrNotFound
}

func TestAuthMiddleware_ValidToken(t *testing.T) {
	repo := &stubRepo{users: make(map[string]*model.User)}
	authSvc := service.NewAuthService(repo, "test-secret", time.Hour)

	token, err := authSvc.Register(context.Background(), "user", "pass")
	require.NoError(t, err)

	var capturedID uuid.UUID
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, ok := middleware.GetUserID(r)
		assert.True(t, ok)
		capturedID = id
		w.WriteHeader(http.StatusOK)
	})

	handler := middleware.Auth(authSvc)(next)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NotEqual(t, uuid.Nil, capturedID)
}

func TestAuthMiddleware_MissingToken(t *testing.T) {
	authSvc := service.NewAuthService(&stubRepo{users: make(map[string]*model.User)}, "secret", time.Hour)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := middleware.Auth(authSvc)(next)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	authSvc := service.NewAuthService(&stubRepo{users: make(map[string]*model.User)}, "secret", time.Hour)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := middleware.Auth(authSvc)(next)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer invalid.token.value")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestGetUserID_NotInContext(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	_, ok := middleware.GetUserID(req)
	assert.False(t, ok)
}

// captureHandler перехватывает slog-записи для проверки в тестах.
type captureHandler struct {
	records []slog.Record
}

func (h *captureHandler) Enabled(_ context.Context, _ slog.Level) bool { return true }
func (h *captureHandler) Handle(_ context.Context, r slog.Record) error {
	h.records = append(h.records, r)
	return nil
}
func (h *captureHandler) WithAttrs(_ []slog.Attr) slog.Handler { return h }
func (h *captureHandler) WithGroup(_ string) slog.Handler      { return h }

// attrValue ищет атрибут по ключу в записи slog.
func attrValue(r slog.Record, key string) (slog.Value, bool) {
	var found slog.Value
	var ok bool
	r.Attrs(func(a slog.Attr) bool {
		if a.Key == key {
			found = a.Value
			ok = true
			return false
		}
		return true
	})
	return found, ok
}

func withCaptureLogger(t *testing.T) *captureHandler {
	t.Helper()
	h := &captureHandler{}
	old := slog.Default()
	slog.SetDefault(slog.New(h))
	t.Cleanup(func() { slog.SetDefault(old) })
	return h
}

func TestLoggerMiddleware_LogsRequest(t *testing.T) {
	cap := withCaptureLogger(t)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	})

	middleware.Logger(next).ServeHTTP(
		httptest.NewRecorder(),
		httptest.NewRequest(http.MethodPost, "/api/items", nil),
	)

	require.Len(t, cap.records, 1)
	rec := cap.records[0]
	assert.Equal(t, "HTTP запрос", rec.Message)

	method, ok := attrValue(rec, "method")
	require.True(t, ok)
	assert.Equal(t, http.MethodPost, method.String())

	path, ok := attrValue(rec, "path")
	require.True(t, ok)
	assert.Equal(t, "/api/items", path.String())

	status, ok := attrValue(rec, "status")
	require.True(t, ok)
	assert.Equal(t, int64(http.StatusCreated), status.Int64())

	_, ok = attrValue(rec, "duration")
	assert.True(t, ok, "поле duration должно присутствовать в логе")
}

func TestLoggerMiddleware_DefaultStatus(t *testing.T) {
	cap := withCaptureLogger(t)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// не вызываем WriteHeader — должен использоваться 200 по умолчанию
		_, _ = w.Write([]byte("ok"))
	})

	middleware.Logger(next).ServeHTTP(
		httptest.NewRecorder(),
		httptest.NewRequest(http.MethodGet, "/health", nil),
	)

	require.Len(t, cap.records, 1)
	status, ok := attrValue(cap.records[0], "status")
	require.True(t, ok)
	assert.Equal(t, int64(http.StatusOK), status.Int64())
}
