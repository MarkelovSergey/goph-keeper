package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad(t *testing.T) {
	tests := []struct {
		name        string
		envs        map[string]string
		wantErr     bool
		errContains string
		check       func(t *testing.T, cfg *Config)
	}{
		{
			name: "все параметры заданы",
			envs: map[string]string{
				"DATABASE_DSN": "postgres://user:pass@localhost/db",
				"JWT_SECRET":   "supersecret",
				"LISTEN_ADDR":  ":9090",
				"TLS_CERT":     "/path/to/cert",
				"TLS_KEY":      "/path/to/key",
			},
			check: func(t *testing.T, cfg *Config) {
				assert.Equal(t, ":9090", cfg.ListenAddr)
				assert.Equal(t, "postgres://user:pass@localhost/db", cfg.DatabaseDSN)
				assert.Equal(t, "supersecret", cfg.JWTSecret)
				assert.Equal(t, "/path/to/cert", cfg.TLSCertPath)
				assert.Equal(t, "/path/to/key", cfg.TLSKeyPath)
			},
		},
		{
			name: "адрес по умолчанию",
			envs: map[string]string{
				"DATABASE_DSN": "postgres://user:pass@localhost/db",
				"JWT_SECRET":   "supersecret",
				"LISTEN_ADDR":  "",
			},
			check: func(t *testing.T, cfg *Config) {
				assert.Equal(t, ":8080", cfg.ListenAddr)
			},
		},
		{
			name: "отсутствует DATABASE_DSN — ошибка",
			envs: map[string]string{
				"DATABASE_DSN": "",
				"JWT_SECRET":   "supersecret",
			},
			wantErr:     true,
			errContains: "DATABASE_DSN",
		},
		{
			name: "отсутствует JWT_SECRET — ошибка",
			envs: map[string]string{
				"DATABASE_DSN": "postgres://user:pass@localhost/db",
				"JWT_SECRET":   "",
			},
			wantErr:     true,
			errContains: "JWT_SECRET",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			for k, v := range tc.envs {
				t.Setenv(k, v)
			}

			cfg, err := Load()
			if tc.wantErr {
				require.Error(t, err)
				if tc.errContains != "" {
					assert.Contains(t, err.Error(), tc.errContains)
				}
				return
			}
			require.NoError(t, err)
			if tc.check != nil {
				tc.check(t, cfg)
			}
		})
	}
}
