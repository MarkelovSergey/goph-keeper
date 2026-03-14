package model

import (
	"time"

	"github.com/google/uuid"
)

// CredentialType — тип хранимых учётных данных.
type CredentialType string

const (
	// CredentialTypeLoginPassword — пара логин/пароль.
	CredentialTypeLoginPassword CredentialType = "login_password"
	// CredentialTypeText — произвольный текст.
	CredentialTypeText CredentialType = "text"
	// CredentialTypeBinary — бинарные данные.
	CredentialTypeBinary CredentialType = "binary"
	// CredentialTypeBankCard — данные банковской карты.
	CredentialTypeBankCard CredentialType = "bank_card"
)

// Credential представляет зашифрованные учётные данные пользователя.
type Credential struct {
	// ID — уникальный идентификатор записи.
	ID uuid.UUID `json:"id"`
	// UserID — идентификатор владельца.
	UserID uuid.UUID `json:"user_id"`
	// Type — тип учётных данных.
	Type CredentialType `json:"type"`
	// Name — пользовательское название записи.
	Name string `json:"name"`
	// Metadata — дополнительные метаданные в открытом виде (необязательно).
	Metadata string `json:"metadata"`
	// Data — зашифрованный blob с содержимым.
	Data []byte `json:"data"`
	// CreatedAt — дата создания.
	CreatedAt time.Time `json:"created_at"`
	// UpdatedAt — дата последнего изменения.
	UpdatedAt time.Time `json:"updated_at"`
}
