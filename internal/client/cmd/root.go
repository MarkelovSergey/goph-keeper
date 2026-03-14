// Package cmd содержит определения cobra-команд CLI-клиента GophKeeper.
package cmd

import (
	"github.com/spf13/cobra"

	"github.com/MarkelovSergey/goph-keeper/internal/client/api"
	"github.com/MarkelovSergey/goph-keeper/internal/client/app"
	clientcfg "github.com/MarkelovSergey/goph-keeper/internal/client/config"
)

var (
	cfg          *clientcfg.Config
	stateManager *app.StateManager
	apiClient    *api.Client
)

// NewRootCmd создаёт и возвращает корневую команду CLI.
func NewRootCmd(version, buildDate string) *cobra.Command {
	var serverAddress string

	root := &cobra.Command{
		Use:   "gophkeeper",
		Short: "GophKeeper — менеджер паролей",
		Long:  "GophKeeper — клиент-серверный менеджер паролей с E2E-шифрованием.",
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			cfg = clientcfg.Load()
			if serverAddress != "" {
				cfg.ServerAddress = serverAddress
			}
			stateManager = app.NewStateManager(cfg.ConfigDir)
			apiClient = api.New(cfg.ServerAddress, cfg.TLSInsecure)
			return nil
		},
		SilenceUsage: true,
	}

	root.PersistentFlags().StringVar(&serverAddress, "server", "", "адрес сервера (переопределяет SERVER_ADDRESS)")

	root.AddCommand(
		newRegisterCmd(),
		newLoginCmd(),
		newAddCmd(),
		newListCmd(),
		newGetCmd(),
		newUpdateCmd(),
		newDeleteCmd(),
		newVersionCmd(version, buildDate),
	)

	return root
}
