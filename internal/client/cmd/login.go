package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/MarkelovSergey/goph-keeper/internal/client/app"
)

func newLoginCmd() *cobra.Command {
	var login, password string

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Авторизоваться на сервере",
		RunE: func(cmd *cobra.Command, _ []string) error {
			token, err := apiClient.Login(cmd.Context(), login, password)
			if err != nil {
				return fmt.Errorf("авторизация не удалась: %w", err)
			}

			// Загружаем текущее состояние, чтобы сохранить соль
			state, err := stateManager.Load()
			if err != nil {
				state = &app.State{}
			}
			state.Token = token

			if err := stateManager.Save(state); err != nil {
				return fmt.Errorf("сохранение состояния: %w", err)
			}

			fmt.Printf("Выполнен вход как '%s'.\n", login)
			return nil
		},
	}

	cmd.Flags().StringVar(&login, "login", "", "логин пользователя (обязательно)")
	cmd.Flags().StringVar(&password, "password", "", "пароль пользователя (обязательно)")
	_ = cmd.MarkFlagRequired("login")
	_ = cmd.MarkFlagRequired("password")

	return cmd
}
