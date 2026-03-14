// Package main — точка входа сервера GophKeeper.
//
// @title       GophKeeper API
// @version     1.0
// @description Менеджер паролей и учётных данных
//
// @securityDefinitions.apikey BearerAuth
// @in                         header
// @name                       Authorization
//
// @host     localhost:8080
// @BasePath /
package main

import (
	"fmt"
	"log/slog"
	"os"

	_ "github.com/MarkelovSergey/goph-keeper/docs"
	"github.com/MarkelovSergey/goph-keeper/internal/server/app"
	"github.com/MarkelovSergey/goph-keeper/internal/server/config"
)

// version и buildDate задаются через ldflags во время сборки.
var (
	version   = "dev"
	buildDate = "unknown"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "version" {
		fmt.Printf("goph-keeper-server %s (собран %s)\n", version, buildDate)
		return
	}

	cfg, err := config.Load()
	if err != nil {
		slog.Error("ошибка конфигурации", "error", err)
		os.Exit(1)
	}

	application, err := app.New(cfg)
	if err != nil {
		slog.Error("ошибка инициализации", "error", err)
		os.Exit(1)
	}

	if err := application.Start(); err != nil {
		slog.Error("ошибка выполнения", "error", err)
		os.Exit(1)
	}
}
