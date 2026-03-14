package cmd

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
	"github.com/MarkelovSergey/goph-keeper/internal/client/app"
	"github.com/MarkelovSergey/goph-keeper/internal/client/crypto"
	"github.com/MarkelovSergey/goph-keeper/internal/model"
)

// setupGlobals инициализирует глобальные переменные пакета для теста.
func setupGlobals(t *testing.T, serverURL string, state *app.State) {
	t.Helper()
	sm := app.NewStateManager(t.TempDir())
	if state != nil {
		require.NoError(t, sm.Save(state))
	}
	stateManager = sm
	apiClient = api.New(serverURL, false)
}

func newSalt(t *testing.T) []byte {
	t.Helper()
	salt, err := crypto.GenerateSalt()
	require.NoError(t, err)
	return salt
}

func encryptJSON(t *testing.T, v any, key []byte) []byte {
	t.Helper()
	plain, err := json.Marshal(v)
	require.NoError(t, err)
	data, err := crypto.Encrypt(plain, key)
	require.NoError(t, err)
	return data
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v) //nolint:errcheck
}

func fakeCred(credType model.CredentialType, data []byte) model.Credential {
	return model.Credential{
		ID:        uuid.New(),
		Type:      credType,
		Name:      "test",
		Data:      data,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// =============================================================================
// version
// =============================================================================

func TestVersionCmd_Output(t *testing.T) {
	cmd := newVersionCmd("1.2.3", "2026-03-14")
	out := captureStdout(t, func() {
		cmd.Run(cmd, nil)
	})
	assert.Contains(t, out, "1.2.3")
	assert.Contains(t, out, "2026-03-14")
}

// =============================================================================
// register
// =============================================================================

func TestRegisterCmd(t *testing.T) {
	tests := []struct {
		name       string
		status     int
		authHeader string
		wantErr    bool
		wantToken  string
	}{
		{
			name:       "успех — токен сохраняется",
			status:     http.StatusOK,
			authHeader: "Bearer reg-token",
			wantToken:  "reg-token",
		},
		{
			name:    "конфликт — ошибка",
			status:  http.StatusConflict,
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method)
				if tc.authHeader != "" {
					w.Header().Set("Authorization", tc.authHeader)
				}
				w.WriteHeader(tc.status)
			}))
			defer srv.Close()
			setupGlobals(t, srv.URL, nil)

			cmd := newRegisterCmd()
			cmd.SetContext(context.Background())
			require.NoError(t, cmd.Flags().Set("login", "alice"))
			require.NoError(t, cmd.Flags().Set("password", "pass"))

			err := cmd.RunE(cmd, nil)
			if tc.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			token, err := stateManager.RequireToken()
			require.NoError(t, err)
			assert.Equal(t, tc.wantToken, token)
		})
	}
}

// =============================================================================
// login
// =============================================================================

func TestLoginCmd(t *testing.T) {
	tests := []struct {
		name       string
		status     int
		authHeader string
		state      *app.State
		wantErr    bool
		wantToken  string
	}{
		{
			name:       "успех — токен сохраняется",
			status:     http.StatusOK,
			authHeader: "Bearer login-token",
			wantToken:  "login-token",
		},
		{
			name:    "неверные данные — ошибка",
			status:  http.StatusUnauthorized,
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tc.authHeader != "" {
					w.Header().Set("Authorization", tc.authHeader)
				}
				w.WriteHeader(tc.status)
			}))
			defer srv.Close()
			setupGlobals(t, srv.URL, &app.State{Salt: newSalt(t)})

			cmd := newLoginCmd()
			cmd.SetContext(context.Background())
			require.NoError(t, cmd.Flags().Set("login", "alice"))
			require.NoError(t, cmd.Flags().Set("password", "pass"))

			err := cmd.RunE(cmd, nil)
			if tc.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			token, err := stateManager.RequireToken()
			require.NoError(t, err)
			assert.Equal(t, tc.wantToken, token)
		})
	}
}

// =============================================================================
// list
// =============================================================================

func TestListCmd(t *testing.T) {
	creds := []model.Credential{
		{ID: uuid.New(), Type: model.CredentialTypeText, Name: "текстовая"},
		{ID: uuid.New(), Type: model.CredentialTypeBankCard, Name: "банковская"},
	}

	tests := []struct {
		name        string
		state       *app.State
		flags       map[string]string
		serverResp  any
		wantErr     bool
		wantErrIs   error
		checkOutput func(t *testing.T, out string)
	}{
		{
			name:      "нет токена — ErrNotLoggedIn",
			wantErr:   true,
			wantErrIs: app.ErrNotLoggedIn,
		},
		{
			name:       "пустой список",
			state:      &app.State{Token: "tok"},
			serverResp: []model.Credential{},
			checkOutput: func(t *testing.T, out string) {
				assert.Contains(t, out, "Записей нет")
			},
		},
		{
			name:       "список с записями",
			state:      &app.State{Token: "tok"},
			serverResp: creds,
			checkOutput: func(t *testing.T, out string) {
				assert.Contains(t, out, "текстовая")
				assert.Contains(t, out, "банковская")
			},
		},
		{
			name:       "фильтр по типу",
			state:      &app.State{Token: "tok"},
			flags:      map[string]string{"type": "text"},
			serverResp: creds,
			checkOutput: func(t *testing.T, out string) {
				assert.Contains(t, out, "текстовая")
				assert.NotContains(t, out, "банковская")
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			url := "http://localhost"
			if tc.serverResp != nil {
				srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					writeJSON(w, tc.serverResp)
				}))
				defer srv.Close()
				url = srv.URL
			}
			setupGlobals(t, url, tc.state)

			cmd := newListCmd()
			cmd.SetContext(context.Background())
			for k, v := range tc.flags {
				require.NoError(t, cmd.Flags().Set(k, v))
			}

			var err error
			out := captureStdout(t, func() {
				err = cmd.RunE(cmd, nil)
			})
			if tc.wantErr {
				require.Error(t, err)
				if tc.wantErrIs != nil {
					require.ErrorIs(t, err, tc.wantErrIs)
				}
				return
			}
			require.NoError(t, err)
			if tc.checkOutput != nil {
				tc.checkOutput(t, out)
			}
		})
	}
}

// =============================================================================
// delete
// =============================================================================

func TestDeleteCmd(t *testing.T) {
	validID := uuid.New().String()

	tests := []struct {
		name         string
		id           string
		state        *app.State
		serverStatus int
		wantErr      bool
		wantErrIs    error
		errContains  string
	}{
		{
			name:      "нет токена — ErrNotLoggedIn",
			id:        validID,
			wantErr:   true,
			wantErrIs: app.ErrNotLoggedIn,
		},
		{
			name:    "невалидный ID — ошибка",
			id:      "not-a-uuid",
			state:   &app.State{Token: "tok"},
			wantErr: true,
		},
		{
			name:         "успех",
			id:           validID,
			state:        &app.State{Token: "tok"},
			serverStatus: http.StatusNoContent,
		},
		{
			name:         "не найдена — ошибка",
			id:           validID,
			state:        &app.State{Token: "tok"},
			serverStatus: http.StatusNotFound,
			wantErr:      true,
			errContains:  "не найдена",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			url := "http://localhost"
			if tc.serverStatus != 0 {
				srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					assert.Equal(t, http.MethodDelete, r.Method)
					w.WriteHeader(tc.serverStatus)
				}))
				defer srv.Close()
				url = srv.URL
			}
			setupGlobals(t, url, tc.state)

			cmd := newDeleteCmd()
			cmd.SetContext(context.Background())
			require.NoError(t, cmd.Flags().Set("id", tc.id))

			err := cmd.RunE(cmd, nil)
			if tc.wantErr {
				require.Error(t, err)
				if tc.wantErrIs != nil {
					require.ErrorIs(t, err, tc.wantErrIs)
				}
				if tc.errContains != "" {
					assert.Contains(t, err.Error(), tc.errContains)
				}
				return
			}
			require.NoError(t, err)
		})
	}
}

// =============================================================================
// get
// =============================================================================

func TestGetCmd(t *testing.T) {
	salt := newSalt(t)
	password := "masterpass"
	key := crypto.DeriveKey(password, salt)
	encTextData := encryptJSON(t, textData{Text: "секрет"}, key)

	credBase := fakeCred(model.CredentialTypeText, nil)
	credBase.Name = "заметка"
	credEncrypted := fakeCred(model.CredentialTypeText, encTextData)
	credWithMeta := fakeCred(model.CredentialTypeText, nil)
	credWithMeta.Metadata = "сайт: example.com"

	tests := []struct {
		name          string
		id            string
		state         *app.State
		flags         map[string]string
		serverHandler http.HandlerFunc
		wantErr       bool
		wantErrIs     error
		errContains   string
		checkOutput   func(t *testing.T, out string)
	}{
		{
			name:      "нет токена — ErrNotLoggedIn",
			id:        uuid.New().String(),
			wantErr:   true,
			wantErrIs: app.ErrNotLoggedIn,
		},
		{
			name:    "невалидный ID — ошибка",
			id:      "bad-uuid",
			state:   &app.State{Token: "tok"},
			wantErr: true,
		},
		{
			name:  "не найдена — ошибка",
			id:    uuid.New().String(),
			state: &app.State{Token: "tok"},
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
			},
			wantErr:     true,
			errContains: "не найдена",
		},
		{
			name:  "без пароля — подсказка об опции --password",
			id:    credBase.ID.String(),
			state: &app.State{Token: "tok"},
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				writeJSON(w, credBase)
			},
			checkOutput: func(t *testing.T, out string) {
				assert.Contains(t, out, "заметка")
				assert.Contains(t, out, "--password")
			},
		},
		{
			name:  "с расшифровкой",
			id:    credEncrypted.ID.String(),
			state: &app.State{Token: "tok", Salt: salt},
			flags: map[string]string{"password": password},
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				writeJSON(w, credEncrypted)
			},
			checkOutput: func(t *testing.T, out string) {
				assert.Contains(t, out, "секрет")
			},
		},
		{
			name:  "с метаданными",
			id:    credWithMeta.ID.String(),
			state: &app.State{Token: "tok"},
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				writeJSON(w, credWithMeta)
			},
			checkOutput: func(t *testing.T, out string) {
				assert.Contains(t, out, "example.com")
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			url := "http://localhost"
			if tc.serverHandler != nil {
				srv := httptest.NewServer(tc.serverHandler)
				defer srv.Close()
				url = srv.URL
			}
			setupGlobals(t, url, tc.state)

			cmd := newGetCmd()
			cmd.SetContext(context.Background())
			require.NoError(t, cmd.Flags().Set("id", tc.id))
			for k, v := range tc.flags {
				require.NoError(t, cmd.Flags().Set(k, v))
			}

			var err error
			out := captureStdout(t, func() {
				err = cmd.RunE(cmd, nil)
			})
			if tc.wantErr {
				require.Error(t, err)
				if tc.wantErrIs != nil {
					require.ErrorIs(t, err, tc.wantErrIs)
				}
				if tc.errContains != "" {
					assert.Contains(t, err.Error(), tc.errContains)
				}
				return
			}
			require.NoError(t, err)
			if tc.checkOutput != nil {
				tc.checkOutput(t, out)
			}
		})
	}
}

// =============================================================================
// add
// =============================================================================

func TestAddCmd(t *testing.T) {
	salt := newSalt(t)

	tests := []struct {
		name         string
		state        *app.State
		flags        map[string]string
		serverStatus int
		wantErr      bool
		wantErrIs    error
		errContains  string
	}{
		{
			name:      "нет токена — ErrNotLoggedIn",
			flags:     map[string]string{"type": "text", "name": "note", "password": "pass"},
			wantErr:   true,
			wantErrIs: app.ErrNotLoggedIn,
		},
		{
			name:        "нет соли — ошибка",
			state:       &app.State{Token: "tok"},
			flags:       map[string]string{"type": "text", "name": "note", "password": "pass", "text": "данные"},
			wantErr:     true,
			errContains: "соль",
		},
		{
			name:         "логин/пароль — успех",
			state:        &app.State{Token: "tok", Salt: salt},
			flags:        map[string]string{"type": "login_password", "name": "GitHub", "password": "masterpass", "username": "alice"},
			serverStatus: http.StatusCreated,
		},
		{
			name:         "текст — успех",
			state:        &app.State{Token: "tok", Salt: salt},
			flags:        map[string]string{"type": "text", "name": "заметка", "password": "masterpass", "text": "секретный текст"},
			serverStatus: http.StatusCreated,
		},
		{
			name:         "банковская карта — успех",
			state:        &app.State{Token: "tok", Salt: salt},
			flags:        map[string]string{"type": "bank_card", "name": "Visa", "password": "masterpass", "number": "4111111111111111", "expiry": "12/26", "cvv": "123", "holder": "Ivan"},
			serverStatus: http.StatusCreated,
		},
		{
			name:    "неизвестный тип — ошибка",
			state:   &app.State{Token: "tok", Salt: salt},
			flags:   map[string]string{"type": "unknown_type", "name": "test", "password": "masterpass"},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			url := "http://localhost"
			if tc.serverStatus != 0 {
				cred := fakeCred(model.CredentialTypeLoginPassword, nil)
				srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(tc.serverStatus)
					writeJSON(w, cred)
				}))
				defer srv.Close()
				url = srv.URL
			}
			setupGlobals(t, url, tc.state)

			cmd := newAddCmd()
			cmd.SetContext(context.Background())
			for k, v := range tc.flags {
				require.NoError(t, cmd.Flags().Set(k, v))
			}

			err := cmd.RunE(cmd, nil)
			if tc.wantErr {
				require.Error(t, err)
				if tc.wantErrIs != nil {
					require.ErrorIs(t, err, tc.wantErrIs)
				}
				if tc.errContains != "" {
					assert.Contains(t, err.Error(), tc.errContains)
				}
				return
			}
			require.NoError(t, err)
		})
	}
}

// =============================================================================
// update
// =============================================================================

func TestUpdateCmd(t *testing.T) {
	salt := newSalt(t)
	password := "masterpass"
	key := crypto.DeriveKey(password, salt)
	encData := encryptJSON(t, textData{Text: "старый"}, key)
	existing := fakeCred(model.CredentialTypeText, encData)

	tests := []struct {
		name          string
		id            string
		state         *app.State
		flags         map[string]string
		serverHandler http.HandlerFunc
		wantErr       bool
		wantErrIs     error
		errContains   string
	}{
		{
			name:      "нет токена — ErrNotLoggedIn",
			id:        uuid.New().String(),
			flags:     map[string]string{"password": "pass"},
			wantErr:   true,
			wantErrIs: app.ErrNotLoggedIn,
		},
		{
			name:    "невалидный ID — ошибка",
			id:      "bad-uuid",
			state:   &app.State{Token: "tok"},
			flags:   map[string]string{"password": "pass"},
			wantErr: true,
		},
		{
			name:  "не найдена — ошибка",
			id:    uuid.New().String(),
			state: &app.State{Token: "tok", Salt: newSalt(t)},
			flags: map[string]string{"password": "pass"},
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
			},
			wantErr:     true,
			errContains: "не найдена",
		},
		{
			name:  "успех",
			id:    existing.ID.String(),
			state: &app.State{Token: "tok", Salt: salt},
			flags: map[string]string{"password": password, "name": "новое имя"},
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				switch r.Method {
				case http.MethodGet:
					writeJSON(w, existing)
				case http.MethodPut:
					updated := existing
					updated.Name = "новое имя"
					writeJSON(w, updated)
				}
			},
		},
		{
			name:  "нет соли — ошибка",
			id:    existing.ID.String(),
			state: &app.State{Token: "tok"},
			flags: map[string]string{"password": "pass"},
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				writeJSON(w, existing)
			},
			wantErr:     true,
			errContains: "соль",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			url := "http://localhost"
			if tc.serverHandler != nil {
				srv := httptest.NewServer(tc.serverHandler)
				defer srv.Close()
				url = srv.URL
			}
			setupGlobals(t, url, tc.state)

			cmd := newUpdateCmd()
			cmd.SetContext(context.Background())
			require.NoError(t, cmd.Flags().Set("id", tc.id))
			for k, v := range tc.flags {
				require.NoError(t, cmd.Flags().Set(k, v))
			}

			err := cmd.RunE(cmd, nil)
			if tc.wantErr {
				require.Error(t, err)
				if tc.wantErrIs != nil {
					require.ErrorIs(t, err, tc.wantErrIs)
				}
				if tc.errContains != "" {
					assert.Contains(t, err.Error(), tc.errContains)
				}
				return
			}
			require.NoError(t, err)
		})
	}
}
