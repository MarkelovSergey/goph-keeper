package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/MarkelovSergey/goph-keeper/internal/model"
)

func TestPrettyType(t *testing.T) {
	tests := []struct {
		credType model.CredentialType
		want     string
	}{
		{model.CredentialTypeLoginPassword, "логин/пароль"},
		{model.CredentialTypeText, "текст"},
		{model.CredentialTypeBinary, "файл"},
		{model.CredentialTypeBankCard, "банк. карта"},
		{"unknown_type", "unknown_type"},
	}

	for _, tc := range tests {
		t.Run(string(tc.credType), func(t *testing.T) {
			assert.Equal(t, tc.want, prettyType(tc.credType))
		})
	}
}

func mustMarshal(t *testing.T, v any) []byte {
	t.Helper()
	b, err := json.Marshal(v)
	require.NoError(t, err)
	return b
}

func TestBuildPlainText(t *testing.T) {
	dir := t.TempDir()
	binFile := filepath.Join(dir, "test.bin")
	binContent := []byte("бинарные данные")
	require.NoError(t, os.WriteFile(binFile, binContent, 0600))

	tests := []struct {
		name        string
		typeStr     string
		username    string
		password    string
		text        string
		file        string
		number      string
		expiry      string
		cvv         string
		holder      string
		wantType    model.CredentialType
		wantErr     bool
		errContains string
		check       func(t *testing.T, data []byte)
	}{
		{
			name:     "логин/пароль",
			typeStr:  "login_password",
			username: "user",
			password: "pass",
			wantType: model.CredentialTypeLoginPassword,
			check: func(t *testing.T, data []byte) {
				var d loginPasswordData
				require.NoError(t, json.Unmarshal(data, &d))
				assert.Equal(t, "user", d.Username)
				assert.Equal(t, "pass", d.Password)
			},
		},
		{
			name:     "текст",
			typeStr:  "text",
			text:     "секретный текст",
			wantType: model.CredentialTypeText,
			check: func(t *testing.T, data []byte) {
				var d textData
				require.NoError(t, json.Unmarshal(data, &d))
				assert.Equal(t, "секретный текст", d.Text)
			},
		},
		{
			name:        "текст пустой — ошибка",
			typeStr:     "text",
			wantErr:     true,
			errContains: "--text",
		},
		{
			name:     "бинарный файл",
			typeStr:  "binary",
			file:     binFile,
			wantType: model.CredentialTypeBinary,
			check: func(t *testing.T, data []byte) {
				var d binaryData
				require.NoError(t, json.Unmarshal(data, &d))
				assert.Equal(t, binFile, d.Filename)
				assert.Equal(t, binContent, d.Content)
			},
		},
		{
			name:        "бинарный без файла — ошибка",
			typeStr:     "binary",
			wantErr:     true,
			errContains: "--file",
		},
		{
			name:    "бинарный несуществующий файл — ошибка",
			typeStr: "binary",
			file:    "/nonexistent/path/file.bin",
			wantErr: true,
		},
		{
			name:     "банковская карта",
			typeStr:  "bank_card",
			number:   "4111111111111111",
			expiry:   "12/26",
			cvv:      "123",
			holder:   "Иван Иванов",
			wantType: model.CredentialTypeBankCard,
			check: func(t *testing.T, data []byte) {
				var d bankCardData
				require.NoError(t, json.Unmarshal(data, &d))
				assert.Equal(t, "4111111111111111", d.Number)
				assert.Equal(t, "12/26", d.Expiry)
				assert.Equal(t, "123", d.CVV)
				assert.Equal(t, "Иван Иванов", d.Holder)
			},
		},
		{
			name:        "неизвестный тип — ошибка",
			typeStr:     "unknown_type",
			wantErr:     true,
			errContains: "unknown_type",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			data, cType, err := buildPlainText(tc.typeStr, tc.username, tc.password, tc.text, tc.file, tc.number, tc.expiry, tc.cvv, tc.holder)
			if tc.wantErr {
				require.Error(t, err)
				if tc.errContains != "" {
					assert.Contains(t, err.Error(), tc.errContains)
				}
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.wantType, cType)
			if tc.check != nil {
				tc.check(t, data)
			}
		})
	}
}

func TestMergeData(t *testing.T) {
	dir := t.TempDir()
	newBinFile := filepath.Join(dir, "new.bin")
	require.NoError(t, os.WriteFile(newBinFile, []byte("new content"), 0600))

	tests := []struct {
		name     string
		credType model.CredentialType
		original []byte
		username string
		password string
		text     string
		file     string
		number   string
		expiry   string
		cvv      string
		holder   string
		wantErr  bool
		check    func(t *testing.T, result []byte)
	}{
		{
			name:     "логин/пароль — обновление обоих полей",
			credType: model.CredentialTypeLoginPassword,
			original: mustMarshal(t, loginPasswordData{Username: "old_user", Password: "old_pass"}),
			username: "new_user",
			password: "new_pass",
			check: func(t *testing.T, result []byte) {
				var d loginPasswordData
				require.NoError(t, json.Unmarshal(result, &d))
				assert.Equal(t, "new_user", d.Username)
				assert.Equal(t, "new_pass", d.Password)
			},
		},
		{
			name:     "логин/пароль — частичное обновление",
			credType: model.CredentialTypeLoginPassword,
			original: mustMarshal(t, loginPasswordData{Username: "old_user", Password: "old_pass"}),
			username: "new_user",
			check: func(t *testing.T, result []byte) {
				var d loginPasswordData
				require.NoError(t, json.Unmarshal(result, &d))
				assert.Equal(t, "new_user", d.Username)
				assert.Equal(t, "old_pass", d.Password)
			},
		},
		{
			name:     "текст — обновление",
			credType: model.CredentialTypeText,
			original: mustMarshal(t, textData{Text: "старый текст"}),
			text:     "новый текст",
			check: func(t *testing.T, result []byte) {
				var d textData
				require.NoError(t, json.Unmarshal(result, &d))
				assert.Equal(t, "новый текст", d.Text)
			},
		},
		{
			name:     "текст — без изменений",
			credType: model.CredentialTypeText,
			original: mustMarshal(t, textData{Text: "текст"}),
			check: func(t *testing.T, result []byte) {
				var d textData
				require.NoError(t, json.Unmarshal(result, &d))
				assert.Equal(t, "текст", d.Text)
			},
		},
		{
			name:     "бинарный — обновление файла",
			credType: model.CredentialTypeBinary,
			original: mustMarshal(t, binaryData{Filename: "old.bin", Content: []byte("old content")}),
			file:     newBinFile,
			check: func(t *testing.T, result []byte) {
				var d binaryData
				require.NoError(t, json.Unmarshal(result, &d))
				assert.Equal(t, newBinFile, d.Filename)
				assert.Equal(t, []byte("new content"), d.Content)
			},
		},
		{
			name:     "бинарный — несуществующий файл — ошибка",
			credType: model.CredentialTypeBinary,
			original: mustMarshal(t, binaryData{Filename: "old.bin", Content: []byte("data")}),
			file:     "/nonexistent/file.bin",
			wantErr:  true,
		},
		{
			name:     "банковская карта — обновление всех полей",
			credType: model.CredentialTypeBankCard,
			original: mustMarshal(t, bankCardData{Number: "0000", Expiry: "01/25", CVV: "000", Holder: "Старый"}),
			number:   "4111111111111111",
			expiry:   "12/26",
			cvv:      "123",
			holder:   "Новый",
			check: func(t *testing.T, result []byte) {
				var d bankCardData
				require.NoError(t, json.Unmarshal(result, &d))
				assert.Equal(t, "4111111111111111", d.Number)
				assert.Equal(t, "12/26", d.Expiry)
				assert.Equal(t, "123", d.CVV)
				assert.Equal(t, "Новый", d.Holder)
			},
		},
		{
			name:     "банковская карта — частичное обновление",
			credType: model.CredentialTypeBankCard,
			original: mustMarshal(t, bankCardData{Number: "0000", Expiry: "01/25", CVV: "000", Holder: "Старый"}),
			number:   "4111111111111111",
			check: func(t *testing.T, result []byte) {
				var d bankCardData
				require.NoError(t, json.Unmarshal(result, &d))
				assert.Equal(t, "4111111111111111", d.Number)
				assert.Equal(t, "01/25", d.Expiry)
				assert.Equal(t, "000", d.CVV)
				assert.Equal(t, "Старый", d.Holder)
			},
		},
		{
			name:     "неизвестный тип — данные без изменений",
			credType: "unknown_type",
			original: []byte(`{"some":"data"}`),
			check: func(t *testing.T, result []byte) {
				assert.Equal(t, []byte(`{"some":"data"}`), result)
			},
		},
		{
			name:     "невалидный JSON — ошибка",
			credType: model.CredentialTypeLoginPassword,
			original: []byte("not-json"),
			wantErr:  true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := mergeData(tc.credType, tc.original, tc.username, tc.password, tc.text, tc.file, tc.number, tc.expiry, tc.cvv, tc.holder)
			if tc.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			if tc.check != nil {
				tc.check(t, result)
			}
		})
	}
}

func TestNewRootCmd(t *testing.T) {
	t.Run("версия команды", func(t *testing.T) {
		root := NewRootCmd("1.2.3", "2026-03-14")
		vCmd, _, err := root.Find([]string{"version"})
		require.NoError(t, err)
		assert.Equal(t, "version", vCmd.Use)
	})

	t.Run("подкоманды зарегистрированы", func(t *testing.T) {
		root := NewRootCmd("0.0.1", "N/A")
		names := make([]string, 0, len(root.Commands()))
		for _, c := range root.Commands() {
			names = append(names, c.Use)
		}
		for _, expected := range []string{"version", "register", "login", "add", "list", "get", "update", "delete"} {
			t.Run(expected, func(t *testing.T) {
				assert.Contains(t, names, expected)
			})
		}
	})
}
