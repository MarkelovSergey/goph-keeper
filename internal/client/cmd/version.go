package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func (a *App) newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Показать версию клиента",
		Run: func(_ *cobra.Command, _ []string) {
			fmt.Printf("gophkeeper %s (собран %s)\n", a.version, a.buildDate)
		},
	}
}
