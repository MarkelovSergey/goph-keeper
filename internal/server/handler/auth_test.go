package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"

	"github.com/MarkelovSergey/goph-keeper/internal/model"
	"github.com/MarkelovSergey/goph-keeper/internal/server/handler"
	"github.com/MarkelovSergey/goph-keeper/internal/server/repository"
	"github.com/MarkelovSergey/goph-keeper/internal/server/service"
)

// stubUserRepo — минимальная заглушка репозитория для хендлер-тестов.
type stubUserRepo struct {
	users map[string]*model.User
}

func newStubUserRepo() *stubUserRepo {
	return &stubUserRepo{users: make(map[string]*model.User)}
}

func (r *stubUserRepo) Create(_ context.Context, user *model.User) error {
	if _, exists := r.users[user.Login]; exists {
		return repository.ErrAlreadyExists
	}
	r.users[user.Login] = user
	return nil
}

func (r *stubUserRepo) GetByLogin(_ context.Context, login string) (*model.User, error) {
	if u, ok := r.users[login]; ok {
		return u, nil
	}
	return nil, repository.ErrNotFound
}

func (r *stubUserRepo) GetByID(_ context.Context, id uuid.UUID) (*model.User, error) {
	for _, u := range r.users {
		if u.ID == id {
			return u, nil
		}
	}
	return nil, repository.ErrNotFound
}

func newTestAuthHandler() *handler.AuthHandler {
	svc := service.NewAuthService(newStubUserRepo(), "test-secret", 24*time.Hour)
	return handler.NewAuthHandler(svc)
}

func TestAuthHandler_Register_Success(t *testing.T) {
	h := newTestAuthHandler()

	body, _ := json.Marshal(map[string]string{"login": "user1", "password": "pass123"})
	req := httptest.NewRequest(http.MethodPost, "/api/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.Register(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NotEmpty(t, w.Header().Get("Authorization"))
}

func TestAuthHandler_Register_MissingFields(t *testing.T) {
	h := newTestAuthHandler()

	body, _ := json.Marshal(map[string]string{"login": "user1"}) // нет пароля
	req := httptest.NewRequest(http.MethodPost, "/api/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.Register(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthHandler_Register_Conflict(t *testing.T) {
	repo := newStubUserRepo()
	hash, _ := bcrypt.GenerateFromPassword([]byte("pass"), bcrypt.DefaultCost)
	repo.users["alice"] = &model.User{ID: uuid.New(), Login: "alice", PasswordHash: string(hash)}

	svc := service.NewAuthService(repo, "secret", time.Hour)
	h := handler.NewAuthHandler(svc)

	body, _ := json.Marshal(map[string]string{"login": "alice", "password": "pass"})
	req := httptest.NewRequest(http.MethodPost, "/api/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.Register(w, req)
	assert.Equal(t, http.StatusConflict, w.Code)
}

func TestAuthHandler_Login_Success(t *testing.T) {
	repo := newStubUserRepo()
	hash, _ := bcrypt.GenerateFromPassword([]byte("correct"), bcrypt.DefaultCost)
	repo.users["bob"] = &model.User{ID: uuid.New(), Login: "bob", PasswordHash: string(hash)}

	svc := service.NewAuthService(repo, "secret", time.Hour)
	h := handler.NewAuthHandler(svc)

	body, _ := json.Marshal(map[string]string{"login": "bob", "password": "correct"})
	req := httptest.NewRequest(http.MethodPost, "/api/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.Login(w, req)
	require.Equal(t, http.StatusOK, w.Code)
	assert.NotEmpty(t, w.Header().Get("Authorization"))
}

func TestAuthHandler_Login_WrongPassword(t *testing.T) {
	repo := newStubUserRepo()
	hash, _ := bcrypt.GenerateFromPassword([]byte("correct"), bcrypt.DefaultCost)
	repo.users["bob"] = &model.User{ID: uuid.New(), Login: "bob", PasswordHash: string(hash)}

	svc := service.NewAuthService(repo, "secret", time.Hour)
	h := handler.NewAuthHandler(svc)

	body, _ := json.Marshal(map[string]string{"login": "bob", "password": "wrong"})
	req := httptest.NewRequest(http.MethodPost, "/api/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.Login(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthHandler_Login_InvalidJSON(t *testing.T) {
	h := newTestAuthHandler()

	req := httptest.NewRequest(http.MethodPost, "/api/login", bytes.NewReader([]byte("not json")))
	w := httptest.NewRecorder()

	h.Login(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// mockAuthService — мок сервиса аутентификации для тестирования ошибок сервиса.
type mockAuthService struct {
	registerErr error
	loginErr    error
	token       string
}

func (m *mockAuthService) Register(_ context.Context, _, _ string) (string, error) {
	return m.token, m.registerErr
}

func (m *mockAuthService) Login(_ context.Context, _, _ string) (string, error) {
	return m.token, m.loginErr
}

func TestAuthHandler_Register_ServiceError(t *testing.T) {
	svc := &mockAuthService{registerErr: errors.New("внутренняя ошибка БД")}
	h := handler.NewAuthHandler(svc)

	body, _ := json.Marshal(map[string]string{"login": "user1", "password": "pass123"})
	req := httptest.NewRequest(http.MethodPost, "/api/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.Register(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
