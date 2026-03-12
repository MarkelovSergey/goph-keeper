// Package main — точка входа сервера GophKeeper.
package main

import (
	"fmt"
	"log"
	"os"

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
		log.Fatalf("ошибка конфигурации: %v", err)
	}

	log.Printf("Сервер GophKeeper %s запускается на %s", version, cfg.ListenAddr)
	// TODO: инициализировать приложение и запустить HTTP-сервер (фаза 3+)
}
