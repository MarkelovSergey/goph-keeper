package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/MarkelovSergey/goph-keeper/internal/model"
	"github.com/MarkelovSergey/goph-keeper/internal/server/repository"
)

type userRepo struct {
	db *pgxpool.Pool
}

// NewUserRepository создаёт репозиторий пользователей на основе PostgreSQL.
func NewUserRepository(db *pgxpool.Pool) repository.UserRepository {
	return &userRepo{db: db}
}

func (r *userRepo) Create(ctx context.Context, user *model.User) error {
	const q = `INSERT INTO users (id, login, password_hash, created_at) VALUES ($1, $2, $3, $4)`
	_, err := r.db.Exec(ctx, q, user.ID, user.Login, user.PasswordHash, user.CreatedAt)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == pgCodeAlreadyExists {
		return repository.ErrAlreadyExists
	}
	if err != nil {
		return fmt.Errorf("user create: %w: %w", repository.ErrInternal, err)
	}
	return nil
}

func (r *userRepo) GetByLogin(ctx context.Context, login string) (*model.User, error) {
	const q = `SELECT id, login, password_hash, created_at FROM users WHERE login = $1`
	user := &model.User{}
	err := r.db.QueryRow(ctx, q, login).Scan(&user.ID, &user.Login, &user.PasswordHash, &user.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, repository.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("user get by login: %w: %w", repository.ErrInternal, err)
	}
	return user, nil
}

func (r *userRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	const q = `SELECT id, login, password_hash, created_at FROM users WHERE id = $1`
	user := &model.User{}
	err := r.db.QueryRow(ctx, q, id).Scan(&user.ID, &user.Login, &user.PasswordHash, &user.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, repository.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("user get by id: %w: %w", repository.ErrInternal, err)
	}
	return user, nil
}
