// Package config предоставляет конфигурацию клиента, загружаемую из переменных окружения
// и локального конфигурационного файла.
package config

import (
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

// Config содержит все параметры конфигурации клиента.
type Config struct {
	// ServerAddress — базовый URL сервера GophKeeper, например "https://localhost:8080".
	ServerAddress string
	// TLSInsecure отключает проверку TLS-сертификата (только для разработки).
	TLSInsecure bool
	// ConfigDir — директория, в которой клиент хранит своё состояние (токены, соль и т.д.).
	ConfigDir string
}

// Load возвращает Config, заполненный из переменных окружения со значениями по умолчанию.
// Сначала пытается загрузить файл .env из рабочей директории (если он есть).
// Адрес сервера может быть переопределён позднее через флаги CLI.
func Load() *Config {
	loadDotEnv()

	return &Config{
		ServerAddress: getEnvOrDefault("SERVER_ADDRESS", "http://localhost:8080"),
		TLSInsecure:   os.Getenv("TLS_INSECURE") == "true",
		ConfigDir:     getEnvOrDefault("GOPHKEEPER_CONFIG_DIR", DefaultConfigDir()),
	}
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

func getEnvOrDefault(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

// DefaultConfigDir возвращает путь к директории конфигурации клиента по умолчанию.
func DefaultConfigDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".gophkeeper"
	}
	return home + "/.gophkeeper"
}
