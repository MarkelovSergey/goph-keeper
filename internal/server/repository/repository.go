// Package repository содержит интерфейсы доступа к данным.
package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/MarkelovSergey/goph-keeper/internal/model"
)

// ErrNotFound возвращается репозиторием, когда запись не найдена.
var ErrNotFound = errors.New("запись не найдена")

// UserRepository — интерфейс репозитория пользователей.
type UserRepository interface {
	// Create сохраняет нового пользователя.
	Create(ctx context.Context, user *model.User) error
	// GetByLogin возвращает пользователя по логину.
	GetByLogin(ctx context.Context, login string) (*model.User, error)
	// GetByID возвращает пользователя по ID.
	GetByID(ctx context.Context, id uuid.UUID) (*model.User, error)
}

// CredentialRepository — интерфейс репозитория учётных данных.
type CredentialRepository interface {
	// Create сохраняет новую запись.
	Create(ctx context.Context, cred *model.Credential) error
	// GetByID возвращает запись по ID, принадлежащую указанному пользователю.
	GetByID(ctx context.Context, id, userID uuid.UUID) (*model.Credential, error)
	// ListByUserID возвращает все записи пользователя.
	ListByUserID(ctx context.Context, userID uuid.UUID) ([]*model.Credential, error)
	// Update обновляет существующую запись.
	Update(ctx context.Context, cred *model.Credential) error
	// Delete удаляет запись по ID.
	Delete(ctx context.Context, id, userID uuid.UUID) error
}
