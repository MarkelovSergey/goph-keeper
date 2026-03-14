// Package service содержит бизнес-логику сервера.
package service

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/MarkelovSergey/goph-keeper/internal/model"
	"github.com/MarkelovSergey/goph-keeper/internal/server/repository"
	pgRepo "github.com/MarkelovSergey/goph-keeper/internal/server/repository/postgres"
)

// Ошибки аутентификации.
var (
	ErrInvalidCredentials = errors.New("неверный логин или пароль")
	ErrUserAlreadyExists  = errors.New("пользователь уже существует")
	ErrInvalidToken       = errors.New("невалидный токен")
)

// Claims — JWT-клеймы с идентификатором пользователя.
type Claims struct {
	UserID uuid.UUID `json:"user_id"`
	jwt.RegisteredClaims
}

// AuthService управляет регистрацией, входом и проверкой JWT.
type AuthService struct {
	userRepo  repository.UserRepository
	jwtSecret []byte
	tokenTTL  time.Duration
}

// NewAuthService создаёт сервис аутентификации.
func NewAuthService(userRepo repository.UserRepository, jwtSecret string, tokenTTL time.Duration) *AuthService {
	return &AuthService{
		userRepo,
		[]byte(jwtSecret),
		tokenTTL,
	}
}

// Register регистрирует нового пользователя и возвращает JWT-токен.
func (s *AuthService) Register(ctx context.Context, login, password string) (string, error) {
	_, err := s.userRepo.GetByLogin(ctx, login)
	if err == nil {
		return "", ErrUserAlreadyExists
	}
	if !errors.Is(err, pgRepo.ErrNotFound) {
		return "", err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	user := &model.User{
		ID:           uuid.New(),
		Login:        login,
		PasswordHash: string(hash),
		CreatedAt:    time.Now(),
	}
	if err := s.userRepo.Create(ctx, user); err != nil {
		return "", err
	}

	return s.generateToken(user.ID)
}

// Login аутентифицирует пользователя и возвращает JWT-токен.
func (s *AuthService) Login(ctx context.Context, login, password string) (string, error) {
	user, err := s.userRepo.GetByLogin(ctx, login)
	if errors.Is(err, pgRepo.ErrNotFound) {
		return "", ErrInvalidCredentials
	}
	if err != nil {
		return "", err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", ErrInvalidCredentials
	}

	return s.generateToken(user.ID)
}

// ParseToken разбирает JWT-токен и возвращает ID пользователя.
func (s *AuthService) ParseToken(tokenString string) (uuid.UUID, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return s.jwtSecret, nil
	})
	if err != nil || !token.Valid {
		return uuid.Nil, ErrInvalidToken
	}
	return claims.UserID, nil
}

func (s *AuthService) generateToken(userID uuid.UUID) (string, error) {
	claims := &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.tokenTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}
