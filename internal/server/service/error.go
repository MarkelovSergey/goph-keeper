package service

import "errors"

// Ошибки аутентификации.
var (
	ErrInvalidCredentials = errors.New("неверный логин или пароль")
	ErrUserAlreadyExists  = errors.New("пользователь уже существует")
	ErrInvalidToken       = errors.New("невалидный токен")
)

// ErrNotFound возвращается, когда запись не найдена.
var ErrNotFound = errors.New("запись не найдена")

// ErrInternal возвращается при внутренней ошибке сервиса.
var ErrInternal = errors.New("внутренняя ошибка сервиса")
