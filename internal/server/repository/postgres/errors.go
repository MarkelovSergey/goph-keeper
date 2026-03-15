// Package postgres реализует репозитории на основе PostgreSQL.
package postgres

import "github.com/MarkelovSergey/goph-keeper/internal/server/repository"

// ErrNotFound — псевдоним repository.ErrNotFound для использования внутри пакета.
var ErrNotFound = repository.ErrNotFound
