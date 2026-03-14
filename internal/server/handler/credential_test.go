package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/MarkelovSergey/goph-keeper/internal/model"
	"github.com/MarkelovSergey/goph-keeper/internal/server/handler"
	"github.com/MarkelovSergey/goph-keeper/internal/server/middleware"
	pgRepo "github.com/MarkelovSergey/goph-keeper/internal/server/repository/postgres"
	"github.com/MarkelovSergey/goph-keeper/internal/server/service"
)

// stubCredentialRepo — заглушка для хендлер-тестов учётных данных.
type stubCredentialRepo struct {
	creds map[uuid.UUID]*model.Credential
}

func newStubCredentialRepo() *stubCredentialRepo {
	return &stubCredentialRepo{creds: make(map[uuid.UUID]*model.Credential)}
}

func (r *stubCredentialRepo) Create(_ context.Context, cred *model.Credential) error {
	r.creds[cred.ID] = cred
	return nil
}

func (r *stubCredentialRepo) GetByID(_ context.Context, id, userID uuid.UUID) (*model.Credential, error) {
	c, ok := r.creds[id]
	if !ok || c.UserID != userID {
		return nil, pgRepo.ErrNotFound
	}
	return c, nil
}

func (r *stubCredentialRepo) ListByUserID(_ context.Context, userID uuid.UUID) ([]*model.Credential, error) {
	var result []*model.Credential
	for _, c := range r.creds {
		if c.UserID == userID {
			result = append(result, c)
		}
	}
	return result, nil
}

func (r *stubCredentialRepo) Update(_ context.Context, cred *model.Credential) error {
	if _, ok := r.creds[cred.ID]; !ok {
		return pgRepo.ErrNotFound
	}
	r.creds[cred.ID] = cred
	return nil
}

func (r *stubCredentialRepo) Delete(_ context.Context, id, userID uuid.UUID) error {
	c, ok := r.creds[id]
	if !ok || c.UserID != userID {
		return pgRepo.ErrNotFound
	}
	delete(r.creds, id)
	return nil
}

// withUserID добавляет userID в контекст запроса (имитирует JWT middleware).
func withUserID(r *http.Request, userID uuid.UUID) *http.Request {
	ctx := context.WithValue(r.Context(), middleware.UserIDKey, userID)
	return r.WithContext(ctx)
}

func newTestCredentialHandler() (*handler.CredentialHandler, *stubCredentialRepo) {
	repo := newStubCredentialRepo()
	svc := service.NewCredentialService(repo)
	return handler.NewCredentialHandler(svc), repo
}

func TestCredentialHandler_Create_Success(t *testing.T) {
	h, _ := newTestCredentialHandler()
	userID := uuid.New()

	body, _ := json.Marshal(map[string]any{
		"type": "text",
		"name": "note",
		"data": []byte("encrypted"),
	})
	req := httptest.NewRequest(http.MethodPost, "/api/credentials", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = withUserID(req, userID)
	w := httptest.NewRecorder()

	h.Create(w, req)

	require.Equal(t, http.StatusCreated, w.Code)

	var cred model.Credential
	require.NoError(t, json.NewDecoder(w.Body).Decode(&cred))
	assert.Equal(t, "note", cred.Name)
	assert.Equal(t, userID, cred.UserID)
}

func TestCredentialHandler_Create_MissingFields(t *testing.T) {
	h, _ := newTestCredentialHandler()
	userID := uuid.New()

	body, _ := json.Marshal(map[string]any{"name": "note"}) // нет type
	req := httptest.NewRequest(http.MethodPost, "/api/credentials", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = withUserID(req, userID)
	w := httptest.NewRecorder()

	h.Create(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCredentialHandler_List_Success(t *testing.T) {
	h, repo := newTestCredentialHandler()
	userID := uuid.New()

	now := time.Now()
	repo.creds[uuid.New()] = &model.Credential{ID: uuid.New(), UserID: userID, Name: "c1", Type: model.CredentialTypeText, CreatedAt: now, UpdatedAt: now}
	repo.creds[uuid.New()] = &model.Credential{ID: uuid.New(), UserID: userID, Name: "c2", Type: model.CredentialTypeText, CreatedAt: now, UpdatedAt: now}
	repo.creds[uuid.New()] = &model.Credential{ID: uuid.New(), UserID: uuid.New(), Name: "c3", Type: model.CredentialTypeText, CreatedAt: now, UpdatedAt: now}

	req := httptest.NewRequest(http.MethodGet, "/api/credentials", nil)
	req = withUserID(req, userID)
	w := httptest.NewRecorder()

	h.List(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var creds []*model.Credential
	require.NoError(t, json.NewDecoder(w.Body).Decode(&creds))
	assert.Len(t, creds, 2)
}

func TestCredentialHandler_Get_Success(t *testing.T) {
	h, repo := newTestCredentialHandler()
	userID := uuid.New()
	credID := uuid.New()

	now := time.Now()
	repo.creds[credID] = &model.Credential{ID: credID, UserID: userID, Name: "secret", Type: model.CredentialTypeText, CreatedAt: now, UpdatedAt: now}

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", credID.String())

	req := httptest.NewRequest(http.MethodGet, "/api/credentials/"+credID.String(), nil)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	req = withUserID(req, userID)
	w := httptest.NewRecorder()

	h.Get(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var cred model.Credential
	require.NoError(t, json.NewDecoder(w.Body).Decode(&cred))
	assert.Equal(t, "secret", cred.Name)
}

func TestCredentialHandler_Get_NotFound(t *testing.T) {
	h, _ := newTestCredentialHandler()
	userID := uuid.New()
	credID := uuid.New()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", credID.String())

	req := httptest.NewRequest(http.MethodGet, "/api/credentials/"+credID.String(), nil)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	req = withUserID(req, userID)
	w := httptest.NewRecorder()

	h.Get(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestCredentialHandler_Delete_Success(t *testing.T) {
	h, repo := newTestCredentialHandler()
	userID := uuid.New()
	credID := uuid.New()

	now := time.Now()
	repo.creds[credID] = &model.Credential{ID: credID, UserID: userID, Name: "del", Type: model.CredentialTypeText, CreatedAt: now, UpdatedAt: now}

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", credID.String())

	req := httptest.NewRequest(http.MethodDelete, "/api/credentials/"+credID.String(), nil)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	req = withUserID(req, userID)
	w := httptest.NewRecorder()

	h.Delete(w, req)
	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestCredentialHandler_Update_NotFound(t *testing.T) {
	h, _ := newTestCredentialHandler()
	userID := uuid.New()
	credID := uuid.New()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", credID.String())

	body, _ := json.Marshal(map[string]any{"name": "new", "data": []byte("d")})
	req := httptest.NewRequest(http.MethodPut, "/api/credentials/"+credID.String(), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	req = withUserID(req, userID)
	w := httptest.NewRecorder()

	h.Update(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestCredentialHandler_Update_Success(t *testing.T) {
	h, repo := newTestCredentialHandler()
	userID := uuid.New()
	credID := uuid.New()

	now := time.Now()
	repo.creds[credID] = &model.Credential{ID: credID, UserID: userID, Name: "old", Type: model.CredentialTypeText, CreatedAt: now, UpdatedAt: now}

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", credID.String())

	body, _ := json.Marshal(map[string]any{"name": "new", "data": []byte("updated")})
	req := httptest.NewRequest(http.MethodPut, "/api/credentials/"+credID.String(), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	req = withUserID(req, userID)
	w := httptest.NewRecorder()

	h.Update(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	var cred model.Credential
	require.NoError(t, json.NewDecoder(w.Body).Decode(&cred))
	assert.Equal(t, "new", cred.Name)
}

func TestCredentialHandler_Create_NoUserID(t *testing.T) {
	h, _ := newTestCredentialHandler()

	body, _ := json.Marshal(map[string]any{"type": "text", "name": "note"})
	req := httptest.NewRequest(http.MethodPost, "/api/credentials", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	// не добавляем userID в контекст
	w := httptest.NewRecorder()

	h.Create(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestCredentialHandler_List_NoUserID(t *testing.T) {
	h, _ := newTestCredentialHandler()

	req := httptest.NewRequest(http.MethodGet, "/api/credentials", nil)
	w := httptest.NewRecorder()

	h.List(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestCredentialHandler_Get_InvalidID(t *testing.T) {
	h, _ := newTestCredentialHandler()
	userID := uuid.New()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "not-a-uuid")

	req := httptest.NewRequest(http.MethodGet, "/api/credentials/not-a-uuid", nil)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	req = withUserID(req, userID)
	w := httptest.NewRecorder()

	h.Get(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCredentialHandler_Delete_NoUserID(t *testing.T) {
	h, _ := newTestCredentialHandler()
	credID := uuid.New()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", credID.String())

	req := httptest.NewRequest(http.MethodDelete, "/api/credentials/"+credID.String(), nil)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	h.Delete(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestCredentialHandler_Delete_NotFound(t *testing.T) {
	h, _ := newTestCredentialHandler()
	userID := uuid.New()
	credID := uuid.New()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", credID.String())

	req := httptest.NewRequest(http.MethodDelete, "/api/credentials/"+credID.String(), nil)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	req = withUserID(req, userID)
	w := httptest.NewRecorder()

	h.Delete(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}
