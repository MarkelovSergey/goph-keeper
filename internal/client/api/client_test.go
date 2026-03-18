package api_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/MarkelovSergey/goph-keeper/internal/client/api"
	"github.com/MarkelovSergey/goph-keeper/internal/model"
)

func TestRegister_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/api/register", r.URL.Path)

		var body map[string]string
		require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		assert.Equal(t, "user", body["login"])

		w.Header().Set("Authorization", "Bearer test-token")
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	client := api.New(srv.URL, false)
	token, err := client.Register(context.Background(), "user", "pass")
	require.NoError(t, err)
	assert.Equal(t, "test-token", token)
}

func TestRegister_Conflict(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusConflict)
	}))
	defer srv.Close()

	client := api.New(srv.URL, false)
	_, err := client.Register(context.Background(), "user", "pass")
	assert.ErrorIs(t, err, api.ErrConflict)
}

func TestLogin_Unauthorized(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer srv.Close()

	client := api.New(srv.URL, false)
	_, err := client.Login(context.Background(), "user", "wrongpass")
	assert.ErrorIs(t, err, api.ErrUnauthorized)
}

func TestCreateCredential_Success(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	credID := uuid.New()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/api/credentials", r.URL.Path)
		assert.Equal(t, "Bearer my-token", r.Header.Get("Authorization"))

		cred := model.Credential{
			ID:        credID,
			Type:      model.CredentialTypeText,
			Name:      "test",
			Metadata:  "",
			Data:      []byte("encrypted"),
			CreatedAt: now,
			UpdatedAt: now,
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		require.NoError(t, json.NewEncoder(w).Encode(cred))
	}))
	defer srv.Close()

	client := api.New(srv.URL, false)
	client.SetToken("my-token")

	cred, err := client.CreateCredential(context.Background(), model.CredentialTypeText, "test", "", []byte("encrypted"))
	require.NoError(t, err)
	assert.Equal(t, credID, cred.ID)
	assert.Equal(t, "test", cred.Name)
}

func TestListCredentials_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/api/credentials", r.URL.Path)

		creds := []*model.Credential{
			{ID: uuid.New(), Name: "cred1", Type: model.CredentialTypeText},
			{ID: uuid.New(), Name: "cred2", Type: model.CredentialTypeLoginPassword},
		}
		w.Header().Set("Content-Type", "application/json")
		require.NoError(t, json.NewEncoder(w).Encode(creds))
	}))
	defer srv.Close()

	client := api.New(srv.URL, false)
	client.SetToken("token")

	creds, err := client.ListCredentials(context.Background())
	require.NoError(t, err)
	assert.Len(t, creds, 2)
}

func TestGetCredential_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	client := api.New(srv.URL, false)
	client.SetToken("token")

	_, err := client.GetCredential(context.Background(), uuid.New())
	assert.ErrorIs(t, err, api.ErrNotFound)
}

func TestDeleteCredential_Success(t *testing.T) {
	credID := uuid.New()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)
		assert.Contains(t, r.URL.Path, credID.String())
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	client := api.New(srv.URL, false)
	client.SetToken("token")

	err := client.DeleteCredential(context.Background(), credID)
	assert.NoError(t, err)
}

func TestUpdateCredential_Success(t *testing.T) {
	credID := uuid.New()
	now := time.Now().UTC().Truncate(time.Second)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method)

		cred := model.Credential{
			ID:        credID,
			Name:      "updated",
			Type:      model.CredentialTypeText,
			UpdatedAt: now,
		}
		w.Header().Set("Content-Type", "application/json")
		require.NoError(t, json.NewEncoder(w).Encode(cred))
	}))
	defer srv.Close()

	client := api.New(srv.URL, false)
	client.SetToken("token")

	cred, err := client.UpdateCredential(context.Background(), credID, "updated", "", []byte("data"))
	require.NoError(t, err)
	assert.Equal(t, "updated", cred.Name)
}

func TestGetCredential_Success(t *testing.T) {
	credID := uuid.New()
	now := time.Now().UTC().Truncate(time.Second)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		cred := model.Credential{
			ID: credID, Name: "found", Type: model.CredentialTypeText,
			CreatedAt: now, UpdatedAt: now,
		}
		w.Header().Set("Content-Type", "application/json")
		require.NoError(t, json.NewEncoder(w).Encode(cred))
	}))
	defer srv.Close()

	client := api.New(srv.URL, false)
	client.SetToken("token")

	cred, err := client.GetCredential(context.Background(), credID)
	require.NoError(t, err)
	assert.Equal(t, "found", cred.Name)
}

func TestDeleteCredential_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	client := api.New(srv.URL, false)
	client.SetToken("token")

	err := client.DeleteCredential(context.Background(), uuid.New())
	assert.ErrorIs(t, err, api.ErrNotFound)
}

func TestUpdateCredential_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	client := api.New(srv.URL, false)
	client.SetToken("token")

	_, err := client.UpdateCredential(context.Background(), uuid.New(), "n", "", nil)
	assert.ErrorIs(t, err, api.ErrNotFound)
}

func TestUpdateCredential_Unauthorized(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer srv.Close()

	client := api.New(srv.URL, false)

	_, err := client.UpdateCredential(context.Background(), uuid.New(), "n", "", nil)
	assert.ErrorIs(t, err, api.ErrUnauthorized)
}

func TestDeleteCredential_Unauthorized(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer srv.Close()

	client := api.New(srv.URL, false)

	err := client.DeleteCredential(context.Background(), uuid.New())
	assert.ErrorIs(t, err, api.ErrUnauthorized)
}

func TestListCredentials_Unauthorized(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer srv.Close()

	client := api.New(srv.URL, false)

	_, err := client.ListCredentials(context.Background())
	assert.ErrorIs(t, err, api.ErrUnauthorized)
}

func TestCreateCredential_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	client := api.New(srv.URL, false)
	client.SetToken("token")

	_, err := client.CreateCredential(context.Background(), model.CredentialTypeText, "n", "", nil)
	assert.Error(t, err)
}
