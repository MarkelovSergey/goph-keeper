package cmd

import (
	"encoding/json"
	"io"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/MarkelovSergey/goph-keeper/internal/client/crypto"
	"github.com/MarkelovSergey/goph-keeper/internal/model"
)

// captureStdout перехватывает вывод в os.Stdout во время вызова fn.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	r, w, err := os.Pipe()
	require.NoError(t, err)

	orig := os.Stdout
	os.Stdout = w

	fn()

	w.Close()
	os.Stdout = orig

	out, err := io.ReadAll(r)
	require.NoError(t, err)
	return string(out)
}

// testKey возвращает фиксированный 32-байтный ключ для тестов.
func testKey() []byte {
	return crypto.DeriveKey("test-password", []byte("0123456789abcdef"))
}

// encrypt шифрует произвольную структуру и возвращает байты.
func encrypt(t *testing.T, v any) []byte {
	t.Helper()
	plain, err := json.Marshal(v)
	require.NoError(t, err)
	data, err := crypto.Encrypt(plain, testKey())
	require.NoError(t, err)
	return data
}

func baseCred(credType model.CredentialType, data []byte) *model.Credential {
	return &model.Credential{
		ID:   uuid.New(),
		Type: credType,
		Name: "test",
		Data: data,
	}
}

func TestPrintDecrypted(t *testing.T) {
	wrongKey := crypto.DeriveKey("wrong-password", []byte("0123456789abcdef"))

	tests := []struct {
		name        string
		cred        func() *model.Credential
		key         []byte
		wantErr     bool
		errContains string
		checkOutput func(t *testing.T, out string)
	}{
		{
			name: "логин/пароль",
			cred: func() *model.Credential {
				return baseCred(model.CredentialTypeLoginPassword, encrypt(t, loginPasswordData{Username: "alice", Password: "s3cr3t"}))
			},
			key: testKey(),
			checkOutput: func(t *testing.T, out string) {
				assert.Contains(t, out, "alice")
				assert.Contains(t, out, "s3cr3t")
			},
		},
		{
			name: "текст",
			cred: func() *model.Credential {
				return baseCred(model.CredentialTypeText, encrypt(t, textData{Text: "секретная заметка"}))
			},
			key: testKey(),
			checkOutput: func(t *testing.T, out string) {
				assert.Contains(t, out, "секретная заметка")
			},
		},
		{
			name: "бинарный файл",
			cred: func() *model.Credential {
				return baseCred(model.CredentialTypeBinary, encrypt(t, binaryData{Filename: "photo.jpg", Content: []byte{0x89, 0x50, 0x4e, 0x47}}))
			},
			key: testKey(),
			checkOutput: func(t *testing.T, out string) {
				assert.Contains(t, out, "photo.jpg")
				assert.Contains(t, out, "4 байт")
			},
		},
		{
			name: "банковская карта",
			cred: func() *model.Credential {
				return baseCred(model.CredentialTypeBankCard, encrypt(t, bankCardData{
					Number: "4111111111111111",
					Expiry: "12/26",
					CVV:    "123",
					Holder: "Ivan Ivanov",
				}))
			},
			key: testKey(),
			checkOutput: func(t *testing.T, out string) {
				assert.Contains(t, out, "4111111111111111")
				assert.Contains(t, out, "12/26")
				assert.Contains(t, out, "123")
				assert.Contains(t, out, "Ivan Ivanov")
			},
		},
		{
			name: "неизвестный тип",
			cred: func() *model.Credential {
				raw := []byte(`{"custom":"value"}`)
				data, err := crypto.Encrypt(raw, testKey())
				require.NoError(t, err)
				return baseCred("custom_type", data)
			},
			key: testKey(),
			checkOutput: func(t *testing.T, out string) {
				assert.Contains(t, out, "custom")
			},
		},
		{
			name: "неверный ключ — ошибка",
			cred: func() *model.Credential {
				return baseCred(model.CredentialTypeText, encrypt(t, textData{Text: "секрет"}))
			},
			key:         wrongKey,
			wantErr:     true,
			errContains: "расшифровка не удалась",
		},
		{
			name: "повреждённые данные — ошибка",
			cred: func() *model.Credential {
				return baseCred(model.CredentialTypeText, []byte("not-valid-ciphertext"))
			},
			key:     testKey(),
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cred := tc.cred()
			var err error
			out := captureStdout(t, func() {
				err = printDecrypted(cred, tc.key)
			})
			if tc.wantErr {
				require.Error(t, err)
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
