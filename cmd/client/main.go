// Package main — точка входа CLI-клиента GophKeeper.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/MarkelovSergey/goph-keeper/internal/client/cmd"
)

// version и buildDate задаются через ldflags во время сборки.
var (
	version   = "dev"
	buildDate = "unknown"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	root := cmd.NewRootCmd(version, buildDate)
	if err := root.ExecuteContext(ctx); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
