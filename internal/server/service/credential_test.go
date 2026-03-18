package service_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/MarkelovSergey/goph-keeper/internal/model"
	"github.com/MarkelovSergey/goph-keeper/internal/server/repository"
	"github.com/MarkelovSergey/goph-keeper/internal/server/service"
)

// mockCredentialRepo — заглушка репозитория учётных данных.
type mockCredentialRepo struct {
	creds map[uuid.UUID]*model.Credential
}

func newMockCredentialRepo() *mockCredentialRepo {
	return &mockCredentialRepo{creds: make(map[uuid.UUID]*model.Credential)}
}

func (m *mockCredentialRepo) Create(_ context.Context, cred *model.Credential) error {
	m.creds[cred.ID] = cred
	return nil
}

func (m *mockCredentialRepo) GetByID(_ context.Context, id, userID uuid.UUID) (*model.Credential, error) {
	c, ok := m.creds[id]
	if !ok || c.UserID != userID {
		return nil, repository.ErrNotFound
	}
	return c, nil
}

func (m *mockCredentialRepo) ListByUserID(_ context.Context, userID uuid.UUID) ([]*model.Credential, error) {
	var result []*model.Credential
	for _, c := range m.creds {
		if c.UserID == userID {
			result = append(result, c)
		}
	}
	return result, nil
}

func (m *mockCredentialRepo) Update(_ context.Context, cred *model.Credential) error {
	if _, ok := m.creds[cred.ID]; !ok {
		return repository.ErrNotFound
	}
	m.creds[cred.ID] = cred
	return nil
}

func (m *mockCredentialRepo) Delete(_ context.Context, id, userID uuid.UUID) error {
	c, ok := m.creds[id]
	if !ok || c.UserID != userID {
		return repository.ErrNotFound
	}
	delete(m.creds, id)
	return nil
}

func newCredentialService() (*service.CredentialService, *mockCredentialRepo) {
	repo := newMockCredentialRepo()
	return service.NewCredentialService(repo), repo
}

func TestCredentialService_Create(t *testing.T) {
	svc, _ := newCredentialService()
	userID := uuid.New()

	cred, err := svc.Create(context.Background(), userID, model.CredentialTypeText, "note", "meta", []byte("encrypted"))
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, cred.ID)
	assert.Equal(t, userID, cred.UserID)
	assert.Equal(t, "note", cred.Name)
}

func TestCredentialService_GetByID_Success(t *testing.T) {
	svc, _ := newCredentialService()
	userID := uuid.New()

	created, err := svc.Create(context.Background(), userID, model.CredentialTypeText, "note", "", []byte("data"))
	require.NoError(t, err)

	got, err := svc.GetByID(context.Background(), created.ID, userID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, got.ID)
}

func TestCredentialService_GetByID_WrongUser(t *testing.T) {
	svc, _ := newCredentialService()
	userID := uuid.New()
	otherUserID := uuid.New()

	created, err := svc.Create(context.Background(), userID, model.CredentialTypeText, "note", "", []byte("data"))
	require.NoError(t, err)

	_, err = svc.GetByID(context.Background(), created.ID, otherUserID)
	assert.ErrorIs(t, err, service.ErrNotFound)
}

func TestCredentialService_ListByUserID(t *testing.T) {
	svc, _ := newCredentialService()
	userID := uuid.New()
	otherID := uuid.New()

	_, _ = svc.Create(context.Background(), userID, model.CredentialTypeText, "n1", "", nil)
	_, _ = svc.Create(context.Background(), userID, model.CredentialTypeLoginPassword, "n2", "", nil)
	_, _ = svc.Create(context.Background(), otherID, model.CredentialTypeText, "n3", "", nil)

	creds, err := svc.ListByUserID(context.Background(), userID)
	require.NoError(t, err)
	assert.Len(t, creds, 2)
}

func TestCredentialService_Update(t *testing.T) {
	svc, _ := newCredentialService()
	userID := uuid.New()

	created, err := svc.Create(context.Background(), userID, model.CredentialTypeText, "old", "", []byte("old"))
	require.NoError(t, err)

	updated, err := svc.Update(context.Background(), created.ID, userID, "new", "meta2", []byte("new"))
	require.NoError(t, err)
	assert.Equal(t, "new", updated.Name)
	assert.Equal(t, "meta2", updated.Metadata)
	assert.Equal(t, []byte("new"), updated.Data)
}

func TestCredentialService_Delete(t *testing.T) {
	svc, _ := newCredentialService()
	userID := uuid.New()

	created, err := svc.Create(context.Background(), userID, model.CredentialTypeText, "note", "", nil)
	require.NoError(t, err)

	err = svc.Delete(context.Background(), created.ID, userID)
	require.NoError(t, err)

	_, err = svc.GetByID(context.Background(), created.ID, userID)
	assert.ErrorIs(t, err, service.ErrNotFound)
}

func TestCredentialService_Delete_WrongUser(t *testing.T) {
	svc, _ := newCredentialService()
	userID := uuid.New()

	created, err := svc.Create(context.Background(), userID, model.CredentialTypeText, "note", "", nil)
	require.NoError(t, err)

	err = svc.Delete(context.Background(), created.ID, uuid.New())
	assert.ErrorIs(t, err, service.ErrNotFound)
}
