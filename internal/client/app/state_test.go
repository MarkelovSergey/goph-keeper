package app

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStateManager_Load(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(t *testing.T, dir string)
		wantErr bool
		check   func(t *testing.T, s *State)
	}{
		{
			name: "нет файла — пустое состояние",
			check: func(t *testing.T, s *State) {
				assert.Empty(t, s.Token)
				assert.Empty(t, s.Salt)
			},
		},
		{
			name: "невалидный JSON — ошибка",
			setup: func(t *testing.T, dir string) {
				require.NoError(t, os.WriteFile(filepath.Join(dir, "state.json"), []byte("not-json{{{"), 0600))
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			sm := NewStateManager(dir)
			if tc.setup != nil {
				tc.setup(t, dir)
			}
			state, err := sm.Load()
			if tc.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			if tc.check != nil {
				tc.check(t, state)
			}
		})
	}
}

func TestStateManager_Save(t *testing.T) {
	tests := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "сохраняет и загружает состояние",
			run: func(t *testing.T) {
				sm := NewStateManager(t.TempDir())
				original := &State{Token: "eyJhbGciOiJIUzI1NiJ9.test", Salt: []byte{1, 2, 3, 4, 5}}
				require.NoError(t, sm.Save(original))
				loaded, err := sm.Load()
				require.NoError(t, err)
				assert.Equal(t, original.Token, loaded.Token)
				assert.Equal(t, original.Salt, loaded.Salt)
			},
		},
		{
			name: "создаёт вложенные директории",
			run: func(t *testing.T) {
				base := t.TempDir()
				nested := filepath.Join(base, "nested", "deep")
				sm := NewStateManager(nested)
				require.NoError(t, sm.Save(&State{Token: "tok"}))
				_, err := os.Stat(filepath.Join(nested, "state.json"))
				require.NoError(t, err)
			},
		},
		{
			name: "устанавливает права 0600",
			run: func(t *testing.T) {
				sm := NewStateManager(t.TempDir())
				require.NoError(t, sm.Save(&State{Token: "tok"}))
				info, err := os.Stat(sm.path)
				require.NoError(t, err)
				assert.Equal(t, os.FileMode(0600), info.Mode().Perm())
			},
		},
		{
			name: "перезаписывает существующий файл",
			run: func(t *testing.T) {
				sm := NewStateManager(t.TempDir())
				require.NoError(t, sm.Save(&State{Token: "first"}))
				require.NoError(t, sm.Save(&State{Token: "second"}))
				loaded, err := sm.Load()
				require.NoError(t, err)
				assert.Equal(t, "second", loaded.Token)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, tc.run)
	}
}

func TestStateManager_RequireToken(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(t *testing.T, sm *StateManager)
		wantErr   bool
		wantErrIs error
		wantToken string
	}{
		{
			name: "успех — возвращает токен",
			setup: func(t *testing.T, sm *StateManager) {
				require.NoError(t, sm.Save(&State{Token: "valid-token"}))
			},
			wantToken: "valid-token",
		},
		{
			name:      "нет файла — ErrNotLoggedIn",
			wantErr:   true,
			wantErrIs: ErrNotLoggedIn,
		},
		{
			name: "пустой токен — ErrNotLoggedIn",
			setup: func(t *testing.T, sm *StateManager) {
				require.NoError(t, sm.Save(&State{Token: ""}))
			},
			wantErr:   true,
			wantErrIs: ErrNotLoggedIn,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			sm := NewStateManager(t.TempDir())
			if tc.setup != nil {
				tc.setup(t, sm)
			}
			token, err := sm.RequireToken()
			if tc.wantErr {
				require.Error(t, err)
				if tc.wantErrIs != nil {
					require.ErrorIs(t, err, tc.wantErrIs)
				}
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.wantToken, token)
		})
	}
}
