package cmd

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/spf13/cobra"

	"github.com/MarkelovSergey/goph-keeper/internal/client/api"
	"github.com/MarkelovSergey/goph-keeper/internal/client/crypto"
	"github.com/MarkelovSergey/goph-keeper/internal/model"
)

func (a *App) newGetCmd() *cobra.Command {
	var (
		id      string
		decrypt bool
	)

	cmd := &cobra.Command{
		Use:   "get",
		Short: "Получить запись по ID",
		RunE: func(cmd *cobra.Command, _ []string) error {
			token, err := a.stateManager.RequireToken()
			if err != nil {
				return err
			}
			a.apiClient.SetToken(token)

			credID, err := uuid.Parse(id)
			if err != nil {
				return fmt.Errorf("неверный формат ID: %w", err)
			}

			cred, err := a.apiClient.GetCredential(cmd.Context(), credID)
			if errors.Is(err, api.ErrNotFound) {
				return fmt.Errorf("запись не найдена")
			}
			if err != nil {
				return fmt.Errorf("получение записи: %w", err)
			}

			fmt.Printf("ID:       %s\n", cred.ID)
			fmt.Printf("Имя:      %s\n", cred.Name)
			fmt.Printf("Тип:      %s\n", prettyType(cred.Type))
			if cred.Metadata != "" {
				fmt.Printf("Метадата: %s\n", cred.Metadata)
			}
			fmt.Printf("Создана:  %s\n", cred.CreatedAt.Format("02.01.2006 15:04"))

			if decrypt {
				state, err := a.stateManager.Load()
				if err != nil || len(state.Salt) == 0 {
					return fmt.Errorf("соль не найдена — невозможно расшифровать")
				}
				password, err := readMasterPassword()
				if err != nil {
					return err
				}
				key := crypto.DeriveKey(password, state.Salt)
				if err := printDecrypted(cred, key); err != nil {
					return err
				}
			} else {
				fmt.Println("(передайте --decrypt для расшифровки содержимого)")
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&id, "id", "", "UUID записи (обязательно)")
	cmd.Flags().BoolVar(&decrypt, "decrypt", false, "расшифровать содержимое (запросит мастер-пароль)")
	_ = cmd.MarkFlagRequired("id")

	return cmd
}

// printDecrypted расшифровывает и выводит содержимое записи.
func printDecrypted(cred *model.Credential, key []byte) error {
	plaintext, err := crypto.Decrypt(cred.Data, key)
	if err != nil {
		return fmt.Errorf("расшифровка не удалась (неверный пароль?): %w", err)
	}

	fmt.Println("--- Содержимое ---")
	switch cred.Type {
	case model.CredentialTypeLoginPassword:
		var d loginPasswordData
		if err := json.Unmarshal(plaintext, &d); err != nil {
			return err
		}
		fmt.Printf("Пользователь: %s\n", d.Username)
		fmt.Printf("Пароль:       %s\n", d.Password)

	case model.CredentialTypeText:
		var d textData
		if err := json.Unmarshal(plaintext, &d); err != nil {
			return err
		}
		fmt.Println(d.Text)

	case model.CredentialTypeBinary:
		var d binaryData
		if err := json.Unmarshal(plaintext, &d); err != nil {
			return err
		}
		fmt.Printf("Файл: %s (%d байт)\n", d.Filename, len(d.Content))

	case model.CredentialTypeBankCard:
		var d bankCardData
		if err := json.Unmarshal(plaintext, &d); err != nil {
			return err
		}
		fmt.Printf("Номер:       %s\n", d.Number)
		fmt.Printf("Срок:        %s\n", d.Expiry)
		fmt.Printf("CVV:         %s\n", d.CVV)
		fmt.Printf("Держатель:   %s\n", d.Holder)

	default:
		fmt.Println(string(plaintext))
	}

	return nil
}
