package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"

	"github.com/MarkelovSergey/goph-keeper/internal/model"
	"github.com/MarkelovSergey/goph-keeper/internal/server/repository"
	"github.com/MarkelovSergey/goph-keeper/internal/server/service"
)

// mockUserRepo — заглушка репозитория пользователей для тестов.
type mockUserRepo struct {
	users map[string]*model.User
}

func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{users: make(map[string]*model.User)}
}

func (m *mockUserRepo) Create(_ context.Context, user *model.User) error {
	if _, exists := m.users[user.Login]; exists {
		return repository.ErrAlreadyExists
	}
	m.users[user.Login] = user
	return nil
}

func (m *mockUserRepo) GetByLogin(_ context.Context, login string) (*model.User, error) {
	if u, ok := m.users[login]; ok {
		return u, nil
	}
	return nil, repository.ErrNotFound
}

func (m *mockUserRepo) GetByID(_ context.Context, id uuid.UUID) (*model.User, error) {
	for _, u := range m.users {
		if u.ID == id {
			return u, nil
		}
	}
	return nil, repository.ErrNotFound
}

func newAuthService() *service.AuthService {
	return service.NewAuthService(newMockUserRepo(), "test-secret-key", 24*time.Hour)
}

func TestRegister_Success(t *testing.T) {
	svc := newAuthService()
	token, err := svc.Register(context.Background(), "user1", "password123")
	require.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestRegister_UserAlreadyExists(t *testing.T) {
	svc := newAuthService()
	_, err := svc.Register(context.Background(), "user1", "pass")
	require.NoError(t, err)

	_, err = svc.Register(context.Background(), "user1", "pass2")
	assert.ErrorIs(t, err, service.ErrUserAlreadyExists)
}

func TestLogin_Success(t *testing.T) {
	repo := newMockUserRepo()
	hash, _ := bcrypt.GenerateFromPassword([]byte("pass"), bcrypt.DefaultCost)
	repo.users["alice"] = &model.User{
		ID:           uuid.New(),
		Login:        "alice",
		PasswordHash: string(hash),
	}

	svc := service.NewAuthService(repo, "secret", 24*time.Hour)
	token, err := svc.Login(context.Background(), "alice", "pass")
	require.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestLogin_InvalidPassword(t *testing.T) {
	svc := newAuthService()
	_, _ = svc.Register(context.Background(), "bob", "correct")

	_, err := svc.Login(context.Background(), "bob", "wrong")
	assert.ErrorIs(t, err, service.ErrInvalidCredentials)
}

func TestLogin_UserNotFound(t *testing.T) {
	svc := newAuthService()
	_, err := svc.Login(context.Background(), "nonexistent", "pass")
	assert.ErrorIs(t, err, service.ErrInvalidCredentials)
}

func TestParseToken_Success(t *testing.T) {
	svc := newAuthService()
	token, err := svc.Register(context.Background(), "carol", "pass")
	require.NoError(t, err)

	userID, err := svc.ParseToken(token)
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, userID)
}

func TestParseToken_Invalid(t *testing.T) {
	svc := newAuthService()
	_, err := svc.ParseToken("invalid.token.here")
	assert.ErrorIs(t, err, service.ErrInvalidToken)
}

func TestParseToken_WrongSecret(t *testing.T) {
	svc1 := service.NewAuthService(newMockUserRepo(), "secret1", time.Hour)
	svc2 := service.NewAuthService(newMockUserRepo(), "secret2", time.Hour)

	token, err := svc1.Register(context.Background(), "dave", "pass")
	require.NoError(t, err)

	_, err = svc2.ParseToken(token)
	assert.ErrorIs(t, err, service.ErrInvalidToken)
}
