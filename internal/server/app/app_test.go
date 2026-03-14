package app

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/MarkelovSergey/goph-keeper/internal/server/config"
)

func TestNew_InvalidDSN(t *testing.T) {
	cfg := &config.Config{
		DatabaseDSN: "postgres://invalid:invalid@localhost:9999/nonexistent",
		JWTSecret:   "secret",
		ListenAddr:  ":0",
	}

	_, err := New(cfg)
	require.Error(t, err)
}
