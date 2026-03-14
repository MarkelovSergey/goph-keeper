// Package config предоставляет конфигурацию сервера, загружаемую из переменных окружения.
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

// Config содержит все параметры конфигурации сервера.
type Config struct {
	// ListenAddr — TCP-адрес для прослушивания, например ":8080".
	ListenAddr string
	// DatabaseDSN — строка подключения к PostgreSQL.
	DatabaseDSN string
	// JWTSecret — HMAC-секрет для подписи JWT-токенов.
	JWTSecret string
	// TLSCertPath — путь к файлу TLS-сертификата (необязательно).
	TLSCertPath string
	// TLSKeyPath — путь к файлу приватного TLS-ключа (необязательно).
	TLSKeyPath string
}

// Load читает конфигурацию из переменных окружения.
// Сначала пытается загрузить файл .env из рабочей директории (если он есть).
// Обязательные переменные: DATABASE_DSN, JWT_SECRET.
// Необязательные переменные: LISTEN_ADDR (по умолчанию :8080), TLS_CERT, TLS_KEY.
func Load() (*Config, error) {
	loadDotEnv()

	cfg := &Config{
		ListenAddr:  getEnvOrDefault("LISTEN_ADDR", ":8080"),
		DatabaseDSN: os.Getenv("DATABASE_DSN"),
		JWTSecret:   os.Getenv("JWT_SECRET"),
		TLSCertPath: os.Getenv("TLS_CERT"),
		TLSKeyPath:  os.Getenv("TLS_KEY"),
	}

	if cfg.DatabaseDSN == "" {
		return nil, fmt.Errorf("переменная окружения DATABASE_DSN обязательна")
	}
	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("переменная окружения JWT_SECRET обязательна")
	}

	return cfg, nil
}

func getEnvOrDefault(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return fmt.Sprintf(":%v", v)
	}

	return defaultVal
}

func loadDotEnv() {
	dir, err := os.Getwd()
	if err != nil {
		return
	}

	for {
		candidate := filepath.Join(dir, ".env")
		if _, err := os.Stat(candidate); err == nil {
			_ = godotenv.Load(candidate)

			return
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return
		}

		dir = parent
	}
}
