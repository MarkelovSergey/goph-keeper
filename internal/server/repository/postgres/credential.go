package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/MarkelovSergey/goph-keeper/internal/model"
	"github.com/MarkelovSergey/goph-keeper/internal/server/repository"
)

type credentialRepo struct {
	db *pgxpool.Pool
}

// NewCredentialRepository создаёт репозиторий учётных данных на основе PostgreSQL.
func NewCredentialRepository(db *pgxpool.Pool) repository.CredentialRepository {
	return &credentialRepo{db: db}
}

func (r *credentialRepo) Create(ctx context.Context, cred *model.Credential) error {
	const q = `
		INSERT INTO credentials (id, user_id, type, name, metadata, data, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	_, err := r.db.Exec(ctx, q,
		cred.ID, cred.UserID, cred.Type, cred.Name,
		cred.Metadata, cred.Data, cred.CreatedAt, cred.UpdatedAt,
	)
	return err
}

func (r *credentialRepo) GetByID(ctx context.Context, id, userID uuid.UUID) (*model.Credential, error) {
	const q = `
		SELECT id, user_id, type, name, metadata, data, created_at, updated_at
		FROM credentials WHERE id = $1 AND user_id = $2`
	cred := &model.Credential{}
	err := r.db.QueryRow(ctx, q, id, userID).Scan(
		&cred.ID, &cred.UserID, &cred.Type, &cred.Name,
		&cred.Metadata, &cred.Data, &cred.CreatedAt, &cred.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return cred, err
}

func (r *credentialRepo) ListByUserID(ctx context.Context, userID uuid.UUID) ([]*model.Credential, error) {
	const q = `
		SELECT id, user_id, type, name, metadata, data, created_at, updated_at
		FROM credentials WHERE user_id = $1 ORDER BY created_at DESC`
	rows, err := r.db.Query(ctx, q, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	creds := make([]*model.Credential, 0)
	for rows.Next() {
		cred := &model.Credential{}
		if err := rows.Scan(
			&cred.ID, &cred.UserID, &cred.Type, &cred.Name,
			&cred.Metadata, &cred.Data, &cred.CreatedAt, &cred.UpdatedAt,
		); err != nil {
			return nil, err
		}
		creds = append(creds, cred)
	}
	return creds, rows.Err()
}

func (r *credentialRepo) Update(ctx context.Context, cred *model.Credential) error {
	const q = `
		UPDATE credentials SET name = $1, metadata = $2, data = $3, updated_at = $4
		WHERE id = $5 AND user_id = $6`
	result, err := r.db.Exec(ctx, q, cred.Name, cred.Metadata, cred.Data, cred.UpdatedAt, cred.ID, cred.UserID)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *credentialRepo) Delete(ctx context.Context, id, userID uuid.UUID) error {
	const q = `DELETE FROM credentials WHERE id = $1 AND user_id = $2`
	result, err := r.db.Exec(ctx, q, id, userID)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
