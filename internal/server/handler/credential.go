package handler

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/MarkelovSergey/goph-keeper/internal/model"
	"github.com/MarkelovSergey/goph-keeper/internal/server/middleware"
	"github.com/MarkelovSergey/goph-keeper/internal/server/service"
)

// credentialService — интерфейс сервиса учётных данных.
type credentialService interface {
	Create(ctx context.Context, userID uuid.UUID, credType model.CredentialType, name, metadata string, data []byte) (*model.Credential, error)
	GetByID(ctx context.Context, id, userID uuid.UUID) (*model.Credential, error)
	ListByUserID(ctx context.Context, userID uuid.UUID) ([]*model.Credential, error)
	Update(ctx context.Context, id, userID uuid.UUID, name, metadata string, data []byte) (*model.Credential, error)
	Delete(ctx context.Context, id, userID uuid.UUID) error
}

// CredentialHandler обрабатывает CRUD-запросы для учётных данных.
type CredentialHandler struct {
	svc credentialService
}

// NewCredentialHandler создаёт обработчик учётных данных.
func NewCredentialHandler(svc credentialService) *CredentialHandler {
	return &CredentialHandler{svc: svc}
}

type createCredentialRequest struct {
	Type     model.CredentialType `json:"type"`
	Name     string               `json:"name"`
	Metadata string               `json:"metadata"`
	Data     []byte               `json:"data"`
}

func (r *createCredentialRequest) UnmarshalJSON(b []byte) error {
	type alias createCredentialRequest
	var a alias
	if err := json.Unmarshal(b, &a); err != nil {
		return err
	}
	if a.Type != "" && !a.Type.IsValid() {
		return errors.New("недопустимый тип учётных данных")
	}

	*r = createCredentialRequest(a)
	return nil
}

type updateCredentialRequest struct {
	Name     string `json:"name"`
	Metadata string `json:"metadata"`
	Data     []byte `json:"data"`
}

// Create godoc
// @Summary  Создать учётные данные
// @Tags     credentials
// @Security BearerAuth
// @Accept   json
// @Produce  json
// @Param    body body     createCredentialRequest true "Учётные данные"
// @Success  201  {object} model.Credential
// @Failure  400  {string} string
// @Failure  401  {string} string
// @Router   /api/credentials [post]
func (h *CredentialHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r)
	if !ok {
		http.Error(w, "ошибка авторизации", http.StatusUnauthorized)
		return
	}

	var req createCredentialRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "неверный формат запроса", http.StatusBadRequest)
		return
	}
	if req.Name == "" || req.Type == "" {
		http.Error(w, "имя и тип обязательны", http.StatusBadRequest)
		return
	}

	cred, err := h.svc.Create(r.Context(), userID, req.Type, req.Name, req.Metadata, req.Data)
	if err != nil {
		slog.Error("создание учётных данных: ошибка сервиса", "error", err)
		http.Error(w, "внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	data, err := json.Marshal(cred)
	if err != nil {
		slog.Error("создание учётных данных: ошибка кодирования ответа", "error", err)
		http.Error(w, "ошибка записи ответа", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_, _ = w.Write(data)
}

// List godoc
// @Summary  Список учётных данных
// @Tags     credentials
// @Security BearerAuth
// @Produce  json
// @Success  200 {array}  model.Credential
// @Failure  401 {string} string
// @Router   /api/credentials [get]
func (h *CredentialHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r)
	if !ok {
		http.Error(w, "ошибка авторизации", http.StatusUnauthorized)
		return
	}

	creds, err := h.svc.ListByUserID(r.Context(), userID)
	if err != nil {
		slog.Error("список учётных данных: ошибка сервиса", "error", err)
		http.Error(w, "внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(creds); err != nil {
		slog.Error("список учётных данных: ошибка кодирования ответа", "error", err)
		http.Error(w, "ошибка записи ответа", http.StatusInternalServerError)
	}
}

// Get godoc
// @Summary  Получить учётные данные по ID
// @Tags     credentials
// @Security BearerAuth
// @Produce  json
// @Param    id  path     string true "ID записи"
// @Success  200 {object} model.Credential
// @Failure  401 {string} string
// @Failure  404 {string} string
// @Router   /api/credentials/{id} [get]
func (h *CredentialHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r)
	if !ok {
		http.Error(w, "ошибка авторизации", http.StatusUnauthorized)
		return
	}

	credID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "неверный формат ID", http.StatusBadRequest)
		return
	}

	cred, err := h.svc.GetByID(r.Context(), credID, userID)
	if errors.Is(err, service.ErrNotFound) {
		http.Error(w, "учётные данные не найдены", http.StatusNotFound)
		return
	}
	if err != nil {
		slog.Error("получение учётных данных: ошибка сервиса", "error", err)
		http.Error(w, "внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(cred); err != nil {
		slog.Error("получение учётных данных: ошибка кодирования ответа", "error", err)
		http.Error(w, "ошибка записи ответа", http.StatusInternalServerError)
	}
}

// Update godoc
// @Summary  Обновить учётные данные
// @Tags     credentials
// @Security BearerAuth
// @Accept   json
// @Produce  json
// @Param    id   path     string                  true "ID записи"
// @Param    body body     updateCredentialRequest true "Новые данные"
// @Success  200  {object} model.Credential
// @Failure  400  {string} string
// @Failure  401  {string} string
// @Failure  404  {string} string
// @Router   /api/credentials/{id} [put]
func (h *CredentialHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r)
	if !ok {
		http.Error(w, "ошибка авторизации", http.StatusUnauthorized)
		return
	}

	credID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "неверный формат ID", http.StatusBadRequest)
		return
	}

	var req updateCredentialRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "неверный формат запроса", http.StatusBadRequest)
		return
	}

	cred, err := h.svc.Update(r.Context(), credID, userID, req.Name, req.Metadata, req.Data)
	if errors.Is(err, service.ErrNotFound) {
		http.Error(w, "учётные данные не найдены", http.StatusNotFound)
		return
	}
	if err != nil {
		slog.Error("обновление учётных данных: ошибка сервиса", "error", err)
		http.Error(w, "внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(cred); err != nil {
		slog.Error("обновление учётных данных: ошибка кодирования ответа", "error", err)
		http.Error(w, "ошибка записи ответа", http.StatusInternalServerError)
	}
}

// Delete godoc
// @Summary  Удалить учётные данные
// @Tags     credentials
// @Security BearerAuth
// @Param    id path string true "ID записи"
// @Success  204
// @Failure  401 {string} string
// @Failure  404 {string} string
// @Router   /api/credentials/{id} [delete]
func (h *CredentialHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r)
	if !ok {
		http.Error(w, "ошибка авторизации", http.StatusUnauthorized)
		return
	}

	credID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "неверный формат ID", http.StatusBadRequest)
		return
	}

	err = h.svc.Delete(r.Context(), credID, userID)
	if errors.Is(err, service.ErrNotFound) {
		http.Error(w, "учётные данные не найдены", http.StatusNotFound)
		return
	}
	if err != nil {
		slog.Error("удаление учётных данных: ошибка сервиса", "error", err)
		http.Error(w, "внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
