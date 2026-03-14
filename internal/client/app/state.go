// Package app содержит логику инициализации и управления состоянием клиента.
package app

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

// ErrNotLoggedIn возвращается, когда токен отсутствует в состоянии.
var ErrNotLoggedIn = errors.New("вы не авторизованы — выполните команду 'login'")

// State хранит локальное состояние клиента: JWT-токен и соль для шифрования.
type State struct {
	// Token — JWT-токен авторизованного пользователя.
	Token string `json:"token"`
	// Salt — соль (hex) для вывода ключа шифрования через Argon2id.
	Salt []byte `json:"salt"`
}

// StateManager управляет сохранением и загрузкой состояния из файла.
type StateManager struct {
	path string
}

// NewStateManager создаёт менеджер состояния, использующий configDir.
func NewStateManager(configDir string) *StateManager {
	return &StateManager{filepath.Join(configDir, "state.json")}
}

// Load читает состояние из файла. Если файл не существует, возвращает пустое состояние.
func (m *StateManager) Load() (*State, error) {
	data, err := os.ReadFile(m.path)
	if errors.Is(err, os.ErrNotExist) {
		return &State{}, nil
	}
	if err != nil {
		return nil, err
	}

	var s State
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, err
	}

	return &s, nil
}

// Save записывает состояние в файл, создавая директорию при необходимости.
func (m *StateManager) Save(s *State) error {
	if err := os.MkdirAll(filepath.Dir(m.path), 0700); err != nil {
		return err
	}

	data, err := json.Marshal(s)
	if err != nil {
		return err
	}

	return os.WriteFile(m.path, data, 0600)
}

// RequireToken возвращает токен из состояния или ошибку ErrNotLoggedIn.
func (m *StateManager) RequireToken() (string, error) {
	s, err := m.Load()
	if err != nil {
		return "", err
	}
	if s.Token == "" {
		return "", ErrNotLoggedIn
	}
	return s.Token, nil
}
