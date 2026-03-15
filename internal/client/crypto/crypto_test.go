package crypto_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/MarkelovSergey/goph-keeper/internal/client/crypto"
)

func TestGenerateSalt(t *testing.T) {
	salt1, err := crypto.GenerateSalt()
	require.NoError(t, err)
	assert.Len(t, salt1, crypto.SaltSize)

	salt2, err := crypto.GenerateSalt()
	require.NoError(t, err)

	// Две соли должны отличаться
	assert.False(t, bytes.Equal(salt1, salt2), "соли должны быть уникальными")
}

func TestDeriveKey(t *testing.T) {
	salt := []byte("testSaltOfLen16B")
	key1 := crypto.DeriveKey("password", salt, nil)
	key2 := crypto.DeriveKey("password", salt, nil)
	key3 := crypto.DeriveKey("other", salt, nil)

	assert.Len(t, key1, crypto.KeySize)
	assert.Equal(t, key1, key2, "одинаковые входные данные должны давать одинаковый ключ")
	assert.NotEqual(t, key1, key3, "разные пароли должны давать разные ключи")
}

func TestEncryptDecrypt(t *testing.T) {
	key := crypto.DeriveKey("secret", []byte("saltSaltSaltSalt"), nil)
	plaintext := []byte("секретные данные для теста")

	ciphertext, err := crypto.Encrypt(plaintext, key)
	require.NoError(t, err)
	assert.NotEqual(t, plaintext, ciphertext)

	decrypted, err := crypto.Decrypt(ciphertext, key)
	require.NoError(t, err)
	assert.Equal(t, plaintext, decrypted)
}

func TestEncryptProducesUniqueOutput(t *testing.T) {
	key := crypto.DeriveKey("secret", []byte("saltSaltSaltSalt"), nil)
	plaintext := []byte("одинаковый текст")

	ct1, err := crypto.Encrypt(plaintext, key)
	require.NoError(t, err)

	ct2, err := crypto.Encrypt(plaintext, key)
	require.NoError(t, err)

	// Каждый вызов должен давать уникальный шифртекст из-за случайного nonce
	assert.False(t, bytes.Equal(ct1, ct2), "шифртексты должны отличаться (разные nonce)")
}

func TestDecryptWrongKey(t *testing.T) {
	key := crypto.DeriveKey("correct", []byte("saltSaltSaltSalt"), nil)
	wrongKey := crypto.DeriveKey("wrong", []byte("saltSaltSaltSalt"), nil)

	ciphertext, err := crypto.Encrypt([]byte("данные"), key)
	require.NoError(t, err)

	_, err = crypto.Decrypt(ciphertext, wrongKey)
	assert.Error(t, err, "расшифровка неверным ключом должна вернуть ошибку")
}

func TestDecryptTooShort(t *testing.T) {
	key := crypto.DeriveKey("secret", []byte("saltSaltSaltSalt"), nil)
	_, err := crypto.Decrypt([]byte("short"), key)
	assert.ErrorIs(t, err, crypto.ErrTooShort)
}

func TestEncryptDecryptEmpty(t *testing.T) {
	key := crypto.DeriveKey("secret", []byte("saltSaltSaltSalt"), nil)
	plaintext := []byte{}

	ciphertext, err := crypto.Encrypt(plaintext, key)
	require.NoError(t, err)

	decrypted, err := crypto.Decrypt(ciphertext, key)
	require.NoError(t, err)
	assert.Empty(t, decrypted)
}
