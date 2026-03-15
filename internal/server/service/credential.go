package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/MarkelovSergey/goph-keeper/internal/model"
	"github.com/MarkelovSergey/goph-keeper/internal/server/repository"
)

// ErrNotFound возвращается, когда запись не найдена.
var ErrNotFound = errors.New("запись не найдена")

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
		return nil, err
	}
	return cred, nil
}

// GetByID возвращает запись по ID.
func (s *CredentialService) GetByID(ctx context.Context, id, userID uuid.UUID) (*model.Credential, error) {
	cred, err := s.repo.GetByID(ctx, id, userID)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, ErrNotFound
	}
	return cred, err
}

// ListByUserID возвращает все записи пользователя.
func (s *CredentialService) ListByUserID(ctx context.Context, userID uuid.UUID) ([]*model.Credential, error) {
	return s.repo.ListByUserID(ctx, userID)
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
		return nil, err
	}
	cred.Name = name
	cred.Metadata = metadata
	cred.Data = data
	cred.UpdatedAt = time.Now()
	if err := s.repo.Update(ctx, cred); err != nil {
		return nil, err
	}
	return cred, nil
}

// Delete удаляет учётные данные.
func (s *CredentialService) Delete(ctx context.Context, id, userID uuid.UUID) error {
	err := s.repo.Delete(ctx, id, userID)
	if errors.Is(err, repository.ErrNotFound) {
		return ErrNotFound
	}
	return err
}
