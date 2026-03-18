// Package model содержит общие модели данных, используемые сервером и клиентом.
package model

import (
	"time"

	"github.com/google/uuid"
)

// User представляет пользователя системы.
type User struct {
	// ID — уникальный идентификатор пользователя.
	ID uuid.UUID `json:"id"`
	// Login — уникальный логин пользователя.
	Login string `json:"login"`
	// PasswordHash — хеш пароля (не передаётся клиенту).
	PasswordHash string `json:"-"`
	// CreatedAt — дата и время регистрации.
	CreatedAt time.Time `json:"created_at"`
}
