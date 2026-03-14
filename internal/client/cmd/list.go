package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/MarkelovSergey/goph-keeper/internal/model"
)

func newListCmd() *cobra.Command {
	var filterType string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "Показать список записей",
		RunE: func(cmd *cobra.Command, _ []string) error {
			token, err := stateManager.RequireToken()
			if err != nil {
				return err
			}
			apiClient.SetToken(token)

			creds, err := apiClient.ListCredentials(cmd.Context())
			if err != nil {
				return fmt.Errorf("получение списка: %w", err)
			}

			if len(creds) == 0 {
				fmt.Println("Записей нет.")
				return nil
			}

			fmt.Printf("%-36s  %-14s  %s\n", "ID", "Тип", "Имя")
			fmt.Println("----------------------------------------------------------------------")
			for _, c := range creds {
				if filterType != "" && string(c.Type) != filterType {
					continue
				}
				fmt.Printf("%-36s  %-14s  %s\n", c.ID, prettyType(c.Type), c.Name)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&filterType, "type", "", "фильтр по типу: login_password, text, binary, bank_card")

	return cmd
}

// prettyType возвращает читаемое название типа.
func prettyType(t model.CredentialType) string {
	switch t {
	case model.CredentialTypeLoginPassword:
		return "логин/пароль"
	case model.CredentialTypeText:
		return "текст"
	case model.CredentialTypeBinary:
		return "файл"
	case model.CredentialTypeBankCard:
		return "банк. карта"
	default:
		return string(t)
	}
}
