package cmd

import (
	"encoding/base64"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/MarkelovSergey/goph-keeper/internal/client/app"
	"github.com/MarkelovSergey/goph-keeper/internal/client/crypto"
)

func (a *App) newRegisterCmd() *cobra.Command {
	var login, password string

	cmd := &cobra.Command{
		Use:   "register",
		Short: "Зарегистрировать нового пользователя",
		RunE: func(cmd *cobra.Command, _ []string) error {
			token, err := a.apiClient.Register(cmd.Context(), login, password)
			if err != nil {
				return fmt.Errorf("регистрация не удалась: %w", err)
			}

			// Генерируем соль для шифрования на стороне клиента
			salt, err := crypto.GenerateSalt()
			if err != nil {
				return fmt.Errorf("генерация соли: %w", err)
			}

			if err := a.stateManager.Save(&app.State{Token: token, Salt: salt, ArgonParams: crypto.DefaultArgonParams()}); err != nil {
				return fmt.Errorf("сохранение состояния: %w", err)
			}

			fmt.Printf("Пользователь '%s' зарегистрирован. Соль: %s\n", login, base64.StdEncoding.EncodeToString(salt))
			return nil
		},
	}

	cmd.Flags().StringVar(&login, "login", "", "логин пользователя (обязательно)")
	cmd.Flags().StringVar(&password, "password", "", "пароль пользователя (обязательно)")
	_ = cmd.MarkFlagRequired("login")
	_ = cmd.MarkFlagRequired("password")

	return cmd
}
