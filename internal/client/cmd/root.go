// Package cmd содержит определения cobra-команд CLI-клиента GophKeeper.
package cmd

import (
	"github.com/spf13/cobra"

	"github.com/MarkelovSergey/goph-keeper/internal/client/api"
	"github.com/MarkelovSergey/goph-keeper/internal/client/app"
	clientcfg "github.com/MarkelovSergey/goph-keeper/internal/client/config"
)

// App содержит зависимости CLI-клиента и используется как ресивер команд.
type App struct {
	version      string
	buildDate    string
	cfg          *clientcfg.Config
	stateManager *app.StateManager
	apiClient    *api.Client
}

// NewRootCmd создаёт и возвращает корневую команду CLI.
func NewRootCmd(version, buildDate string) *cobra.Command {
	var serverAddress string
	a := &App{
		version:   version,
		buildDate: buildDate,
	}

	root := &cobra.Command{
		Use:   "gophkeeper",
		Short: "GophKeeper — менеджер паролей",
		Long:  "GophKeeper — клиент-серверный менеджер паролей с E2E-шифрованием.",
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			a.cfg = clientcfg.Load()
			if serverAddress != "" {
				a.cfg.ServerAddress = serverAddress
			}
			a.stateManager = app.NewStateManager(a.cfg.ConfigDir)
			a.apiClient = api.New(a.cfg.ServerAddress, a.cfg.TLSInsecure)
			return nil
		},
		SilenceUsage: true,
	}

	root.PersistentFlags().StringVar(&serverAddress, "server", "", "адрес сервера (переопределяет SERVER_ADDRESS)")

	root.AddCommand(
		a.newRegisterCmd(),
		a.newLoginCmd(),
		a.newAddCmd(),
		a.newListCmd(),
		a.newGetCmd(),
		a.newUpdateCmd(),
		a.newDeleteCmd(),
		a.newVersionCmd(),
	)

	return root
}
