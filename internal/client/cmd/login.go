package cmd

import (
	"encoding/base64"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/MarkelovSergey/goph-keeper/internal/client/app"
	"github.com/MarkelovSergey/goph-keeper/internal/client/crypto"
)

func (a *App) newLoginCmd() *cobra.Command {
	var login, password string

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Авторизоваться на сервере",
		RunE: func(cmd *cobra.Command, _ []string) error {
			token, err := a.apiClient.Login(cmd.Context(), login, password)
			if err != nil {
				return fmt.Errorf("авторизация не удалась: %w", err)
			}

			// Загружаем текущее состояние, чтобы сохранить соль
			state, err := a.stateManager.Load()
			if err != nil {
				state = &app.State{}
			}
			state.Token = token

			// Генерируем соль, если она отсутствует (первый вход без регистрации через CLI)
			if len(state.Salt) == 0 {
				salt, err := crypto.GenerateSalt()
				if err != nil {
					return fmt.Errorf("генерация соли: %w", err)
				}
				state.Salt = salt
				state.ArgonParams = crypto.DefaultArgonParams()
				fmt.Printf("Сгенерирована новая соль: %s\n", base64.StdEncoding.EncodeToString(salt))
			}

			if err := a.stateManager.Save(state); err != nil {
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
