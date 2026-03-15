package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/MarkelovSergey/goph-keeper/internal/model"
	"github.com/MarkelovSergey/goph-keeper/internal/server/repository"
)

// CredentialService управляет учётными данными пользователей.
type CredentialService struct {
	repo repository.CredentialRepository
}

// NewCredentialService создаёт сервис учётных данных.
func NewCredentialService(repo repository.CredentialRepository) *CredentialService {
	return &CredentialService{repo: repo}
}

// Create сохраняет новые учётные данные.
func (s *CredentialService) Create(
	ctx context.Context,
	userID uuid.UUID,
	credType model.CredentialType,
	name, metadata string,
	data []byte,
) (*model.Credential, error) {
	now := time.Now()
	cred := &model.Credential{
		ID:        uuid.New(),
		UserID:    userID,
		Type:      credType,
		Name:      name,
		Metadata:  metadata,
		Data:      data,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := s.repo.Create(ctx, cred); err != nil {
		return nil, fmt.Errorf("credential create: %w: %w", ErrInternal, err)
	}
	return cred, nil
}

// GetByID возвращает запись по ID.
func (s *CredentialService) GetByID(ctx context.Context, id, userID uuid.UUID) (*model.Credential, error) {
	cred, err := s.repo.GetByID(ctx, id, userID)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("credential get: %w: %w", ErrInternal, err)
	}
	return cred, nil
}

// ListByUserID возвращает все записи пользователя.
func (s *CredentialService) ListByUserID(ctx context.Context, userID uuid.UUID) ([]*model.Credential, error) {
	creds, err := s.repo.ListByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("credential list: %w: %w", ErrInternal, err)
	}
	return creds, nil
}

// Update обновляет существующие учётные данные.
func (s *CredentialService) Update(
	ctx context.Context,
	id, userID uuid.UUID,
	name, metadata string,
	data []byte,
) (*model.Credential, error) {
	cred, err := s.repo.GetByID(ctx, id, userID)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("credential update get: %w: %w", ErrInternal, err)
	}
	cred.Name = name
	cred.Metadata = metadata
	cred.Data = data
	cred.UpdatedAt = time.Now()
	if err := s.repo.Update(ctx, cred); err != nil {
		return nil, fmt.Errorf("credential update: %w: %w", ErrInternal, err)
	}
	return cred, nil
}

// Delete удаляет учётные данные.
func (s *CredentialService) Delete(ctx context.Context, id, userID uuid.UUID) error {
	err := s.repo.Delete(ctx, id, userID)
	if errors.Is(err, repository.ErrNotFound) {
		return ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("credential delete: %w: %w", ErrInternal, err)
	}
	return nil
}
