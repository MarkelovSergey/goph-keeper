package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/google/uuid"
	"github.com/spf13/cobra"

	"github.com/MarkelovSergey/goph-keeper/internal/client/api"
	"github.com/MarkelovSergey/goph-keeper/internal/client/crypto"
	"github.com/MarkelovSergey/goph-keeper/internal/model"
)

func (a *App) newUpdateCmd() *cobra.Command {
	var (
		id          string
		name        string
		metadata    string
		username    string
		newPassword string
		text        string
		filePath    string
		cardNumber  string
		cardExpiry  string
		cardCVV     string
		cardHolder  string
	)

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Обновить существующую запись",
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

			// Получаем текущую запись
			existing, err := a.apiClient.GetCredential(cmd.Context(), credID)
			if errors.Is(err, api.ErrNotFound) {
				return fmt.Errorf("запись не найдена")
			}
			if err != nil {
				return fmt.Errorf("получение записи: %w", err)
			}

			// Получаем ключ шифрования
			state, err := a.stateManager.Load()
			if err != nil || len(state.Salt) == 0 {
				return fmt.Errorf("соль не найдена — невозможно зашифровать данные")
			}

			password, err := readMasterPassword()
			if err != nil {
				return err
			}

			key := crypto.DeriveKey(password, state.Salt)

			// Расшифровываем текущие данные для merge
			var encryptedData []byte
			if len(existing.Data) > 0 {
				decrypted, err := crypto.Decrypt(existing.Data, key)
				if err != nil {
					return fmt.Errorf("расшифровка текущих данных не удалась (неверный пароль?): %w", err)
				}
				updated, err := mergeData(existing.Type, decrypted, username, newPassword, text, filePath, cardNumber, cardExpiry, cardCVV, cardHolder)
				if err != nil {
					return err
				}
				encryptedData, err = crypto.Encrypt(updated, key)
				if err != nil {
					return fmt.Errorf("шифрование: %w", err)
				}
			}

			if name == "" {
				name = existing.Name
			}
			if metadata == "" {
				metadata = existing.Metadata
			}

			cred, err := a.apiClient.UpdateCredential(cmd.Context(), credID, name, metadata, encryptedData)
			if errors.Is(err, api.ErrNotFound) {
				return fmt.Errorf("запись не найдена")
			}
			if err != nil {
				return fmt.Errorf("обновление записи: %w", err)
			}

			fmt.Printf("Запись обновлена: ID=%s, имя=%s\n", cred.ID, cred.Name)
			return nil
		},
	}

	cmd.Flags().StringVar(&id, "id", "", "UUID записи (обязательно)")
	cmd.Flags().StringVar(&name, "name", "", "новое название записи")
	cmd.Flags().StringVar(&metadata, "meta", "", "новые метаданные")
	cmd.Flags().StringVar(&username, "username", "", "имя пользователя (для login_password)")
	cmd.Flags().StringVar(&newPassword, "new-password", "", "новый пароль (для login_password)")
	cmd.Flags().StringVar(&text, "text", "", "текстовое содержимое (для text)")
	cmd.Flags().StringVar(&filePath, "file", "", "путь к файлу (для binary)")
	cmd.Flags().StringVar(&cardNumber, "number", "", "номер карты (для bank_card)")
	cmd.Flags().StringVar(&cardExpiry, "expiry", "", "срок действия (для bank_card)")
	cmd.Flags().StringVar(&cardCVV, "cvv", "", "CVV-код (для bank_card)")
	cmd.Flags().StringVar(&cardHolder, "holder", "", "имя держателя (для bank_card)")

	_ = cmd.MarkFlagRequired("id")

	return cmd
}

// mergeData обновляет только переданные поля, оставляя остальные из decrypted.
func mergeData(credType model.CredentialType, decrypted []byte, username, password, text, filePath, cardNumber, cardExpiry, cardCVV, cardHolder string) ([]byte, error) {
	switch credType {
	case model.CredentialTypeLoginPassword:
		var d loginPasswordData
		if err := json.Unmarshal(decrypted, &d); err != nil {
			return nil, err
		}
		if username != "" {
			d.Username = username
		}
		if password != "" {
			d.Password = password
		}
		return json.Marshal(d)

	case model.CredentialTypeText:
		var d textData
		if err := json.Unmarshal(decrypted, &d); err != nil {
			return nil, err
		}
		if text != "" {
			d.Text = text
		}
		return json.Marshal(d)

	case model.CredentialTypeBinary:
		var d binaryData
		if err := json.Unmarshal(decrypted, &d); err != nil {
			return nil, err
		}
		if filePath != "" {
			content, err := os.ReadFile(filePath)
			if err != nil {
				return nil, fmt.Errorf("чтение файла: %w", err)
			}
			d.Filename = filePath
			d.Content = content
		}
		return json.Marshal(d)

	case model.CredentialTypeBankCard:
		var d bankCardData
		if err := json.Unmarshal(decrypted, &d); err != nil {
			return nil, err
		}
		if cardNumber != "" {
			d.Number = cardNumber
		}
		if cardExpiry != "" {
			d.Expiry = cardExpiry
		}
		if cardCVV != "" {
			d.CVV = cardCVV
		}
		if cardHolder != "" {
			d.Holder = cardHolder
		}
		return json.Marshal(d)

	default:
		return decrypted, nil
	}
}
