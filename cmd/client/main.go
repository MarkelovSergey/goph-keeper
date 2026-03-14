// Package main — точка входа CLI-клиента GophKeeper.
package main

import (
	"fmt"
	"os"

	"github.com/MarkelovSergey/goph-keeper/internal/client/cmd"
)

// version и buildDate задаются через ldflags во время сборки.
var (
	version   = "dev"
	buildDate = "unknown"
)

func main() {
	root := cmd.NewRootCmd(version, buildDate)
	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
