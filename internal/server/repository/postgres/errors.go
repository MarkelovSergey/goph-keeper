// Package postgres реализует репозитории на основе PostgreSQL.
package postgres

import "errors"

// ErrNotFound возвращается, когда запись не найдена в базе данных.
var ErrNotFound = errors.New("запись не найдена")
