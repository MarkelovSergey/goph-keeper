package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newVersionCmd(version, buildDate string) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Показать версию клиента",
		Run: func(_ *cobra.Command, _ []string) {
			fmt.Printf("gophkeeper %s (собран %s)\n", version, buildDate)
		},
	}
}
