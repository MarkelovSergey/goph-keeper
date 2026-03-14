package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	pgRepo "github.com/MarkelovSergey/goph-keeper/internal/server/repository/postgres"
	"github.com/MarkelovSergey/goph-keeper/internal/server/middleware"
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
	return nil, pgRepo.ErrNotFound
}

func (r *stubRepo) GetByID(_ context.Context, id uuid.UUID) (*model.User, error) {
	for _, u := range r.users {
		if u.ID == id {
			return u, nil
		}
	}
	return nil, pgRepo.ErrNotFound
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

func TestLoggerMiddleware(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	})

	handler := middleware.Logger(next)
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestLoggerMiddleware_DefaultStatus(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// не вызываем WriteHeader — должен использоваться 200 по умолчанию
		_, _ = w.Write([]byte("ok"))
	})

	handler := middleware.Logger(next)
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}
