// Package main — точка входа CLI-клиента GophKeeper.
package main

import (
	"fmt"
	"os"
)

// version и buildDate задаются через ldflags во время сборки.
var (
	version   = "dev"
	buildDate = "unknown"
)

func main() {
	// TODO: инициализировать корневую команду cobra и запустить (фаза 6)
	// Пока что выводим версию, если запрошена.
	if len(os.Args) > 1 && os.Args[1] == "version" {
		fmt.Printf("goph-keeper %s (собран %s)\n", version, buildDate)
		os.Exit(0)
	}

	fmt.Fprintf(os.Stderr, "goph-keeper %s — используйте --help для справки\n", version)
}
