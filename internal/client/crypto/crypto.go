// Package crypto реализует E2E-шифрование данных клиента.
// Используется Argon2id для вывода ключа из пароля и AES-256-GCM для шифрования.
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"io"

	"golang.org/x/crypto/argon2"
)

const (
	// KeySize — размер AES-ключа в байтах (256 бит).
	KeySize = 32
	// SaltSize — размер соли для Argon2id.
	SaltSize = 16
)

// ErrTooShort возникает, если шифртекст слишком короток для извлечения nonce.
var ErrTooShort = errors.New("шифртекст слишком короткий")

// ArgonParams хранит параметры алгоритма Argon2id для вывода ключа.
// Параметры сохраняются рядом с солью, чтобы расшифровка оставалась
// возможной при смене значений по умолчанию в будущих версиях.
type ArgonParams struct {
	Time    uint32 `json:"time"`
	Memory  uint32 `json:"memory"`
	Threads uint8  `json:"threads"`
}

// DefaultArgonParams возвращает рекомендуемые параметры Argon2id.
func DefaultArgonParams() *ArgonParams {
	return &ArgonParams{
		Time:    1,
		Memory:  64 * 1024,
		Threads: 4,
	}
}

// GenerateSalt генерирует случайную соль для Argon2id.
func GenerateSalt() ([]byte, error) {
	salt := make([]byte, SaltSize)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return nil, err
	}
	return salt, nil
}

// DeriveKey выводит 32-байтный AES-ключ из пароля и соли с помощью Argon2id.
// Если params равен nil, используются параметры по умолчанию.
func DeriveKey(password string, salt []byte, params *ArgonParams) []byte {
	if params == nil {
		params = DefaultArgonParams()
	}
	return argon2.IDKey([]byte(password), salt, params.Time, params.Memory, params.Threads, KeySize)
}

// Encrypt шифрует plaintext ключом key (AES-256-GCM).
// Результат: nonce (12 байт) || ciphertext.
func Encrypt(plaintext, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}

// Decrypt расшифровывает данные, зашифрованные функцией Encrypt.
// Ожидает формат: nonce (12 байт) || ciphertext.
func Decrypt(ciphertext, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, ErrTooShort
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}
