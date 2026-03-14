package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/MarkelovSergey/goph-keeper/internal/model"
	"github.com/MarkelovSergey/goph-keeper/internal/server/middleware"
	pgRepo "github.com/MarkelovSergey/goph-keeper/internal/server/repository/postgres"
	"github.com/MarkelovSergey/goph-keeper/internal/server/service"
)

// CredentialHandler обрабатывает CRUD-запросы для учётных данных.
type CredentialHandler struct {
	svc *service.CredentialService
}

// NewCredentialHandler создаёт обработчик учётных данных.
func NewCredentialHandler(svc *service.CredentialService) *CredentialHandler {
	return &CredentialHandler{svc}
}

type createCredentialRequest struct {
	Type     model.CredentialType `json:"type"`
	Name     string               `json:"name"`
	Metadata string               `json:"metadata"`
	Data     []byte               `json:"data"`
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
		http.Error(w, "внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(cred)
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
		http.Error(w, "внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}
	if creds == nil {
		creds = []*model.Credential{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(creds)
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
	if errors.Is(err, pgRepo.ErrNotFound) {
		http.Error(w, "учётные данные не найдены", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cred)
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
	if errors.Is(err, pgRepo.ErrNotFound) {
		http.Error(w, "учётные данные не найдены", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cred)
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
	if errors.Is(err, pgRepo.ErrNotFound) {
		http.Error(w, "учётные данные не найдены", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
