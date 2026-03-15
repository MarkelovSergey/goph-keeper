package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoad(t *testing.T) {
	tests := []struct {
		name          string
		serverAddress string
		tlsInsecure   string
		configDir     string
		wantAddress   string
		wantInsecure  bool
		wantConfigDir string // если пусто — проверяем NotEmpty и Contains(".gophkeeper")
	}{
		{
			name:         "значения по умолчанию",
			wantAddress:  "http://localhost:8080",
			wantInsecure: false,
		},
		{
			name:          "переменные окружения",
			serverAddress: "https://example.com:9443",
			tlsInsecure:   "true",
			configDir:     "/tmp/testconfig",
			wantAddress:   "https://example.com:9443",
			wantInsecure:  true,
			wantConfigDir: "/tmp/testconfig",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("SERVER_ADDRESS", tc.serverAddress)
			t.Setenv("TLS_INSECURE", tc.tlsInsecure)
			t.Setenv("GOPHKEEPER_CONFIG_DIR", tc.configDir)

			cfg := Load()

			assert.Equal(t, tc.wantAddress, cfg.ServerAddress)
			assert.Equal(t, tc.wantInsecure, cfg.TLSInsecure)
			if tc.wantConfigDir != "" {
				assert.Equal(t, tc.wantConfigDir, cfg.ConfigDir)
			} else {
				assert.NotEmpty(t, cfg.ConfigDir)
			}
		})
	}
}
