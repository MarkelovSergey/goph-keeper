// Package handler содержит HTTP-обработчики сервера.
package handler

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/MarkelovSergey/goph-keeper/internal/server/service"
)

// authService — интерфейс сервиса аутентификации.
type authService interface {
	Register(ctx context.Context, login, password string) (string, error)
	Login(ctx context.Context, login, password string) (string, error)
}

// AuthHandler обрабатывает запросы регистрации и входа.
type AuthHandler struct {
	authSvc authService
}

// NewAuthHandler создаёт обработчик аутентификации.
func NewAuthHandler(authSvc authService) *AuthHandler {
	return &AuthHandler{authSvc: authSvc}
}

type authRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

// Register godoc
// @Summary Регистрация пользователя
// @Tags    auth
// @Accept  json
// @Produce json
// @Param   body body     authRequest true "Данные для регистрации"
// @Success 200  {string} string
// @Failure 400  {string} string
// @Failure 409  {string} string
// @Router  /api/register [post]
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req authRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "неверный формат запроса", http.StatusBadRequest)
		return
	}
	if req.Login == "" || req.Password == "" {
		http.Error(w, "логин и пароль обязательны", http.StatusBadRequest)
		return
	}

	token, err := h.authSvc.Register(r.Context(), req.Login, req.Password)
	if errors.Is(err, service.ErrUserAlreadyExists) {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}
	if err != nil {
		slog.Error("регистрация пользователя: ошибка сервиса", "error", err)
		http.Error(w, "внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Authorization", "Bearer "+token)
	w.WriteHeader(http.StatusOK)
}

// Login godoc
// @Summary Вход пользователя
// @Tags    auth
// @Accept  json
// @Produce json
// @Param   body body     authRequest true "Данные для входа"
// @Success 200  {string} string
// @Failure 400  {string} string
// @Failure 401  {string} string
// @Router  /api/login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req authRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "неверный формат запроса", http.StatusBadRequest)
		return
	}
	if req.Login == "" || req.Password == "" {
		http.Error(w, "логин и пароль обязательны", http.StatusBadRequest)
		return
	}

	token, err := h.authSvc.Login(r.Context(), req.Login, req.Password)
	if errors.Is(err, service.ErrInvalidCredentials) {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	if err != nil {
		slog.Error("вход пользователя: ошибка сервиса", "error", err)
		http.Error(w, "внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Authorization", "Bearer "+token)
	w.WriteHeader(http.StatusOK)
}
