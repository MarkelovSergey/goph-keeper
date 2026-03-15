package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/MarkelovSergey/goph-keeper/internal/client/crypto"
	"github.com/MarkelovSergey/goph-keeper/internal/model"
)

// loginPasswordData — структура для типа login_password.
type loginPasswordData struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// textData — структура для типа text.
type textData struct {
	Text string `json:"text"`
}

// binaryData — структура для типа binary.
type binaryData struct {
	Filename string `json:"filename"`
	Content  []byte `json:"content"`
}

// bankCardData — структура для типа bank_card.
type bankCardData struct {
	Number string `json:"number"`
	Expiry string `json:"expiry"`
	CVV    string `json:"cvv"`
	Holder string `json:"holder"`
}

func (a *App) newAddCmd() *cobra.Command {
	var (
		credType string
		name     string
		metadata string
		// login_password
		username string
		// text
		text string
		// binary
		filePath string
		// bank_card
		cardNumber string
		cardExpiry string
		cardCVV    string
		cardHolder string
	)

	cmd := &cobra.Command{
		Use:   "add",
		Short: "Добавить новую запись",
		Long: `Добавить новую зашифрованную запись.

Примеры:
  gophkeeper add --type login_password --name "GitHub" --username user --password pass
  gophkeeper add --type text --name "Заметка" --text "секретный текст"
  gophkeeper add --type binary --name "Файл" --file /path/to/file
  gophkeeper add --type bank_card --name "Visa" --number 4111111111111111 --expiry 12/26 --cvv 123 --holder "Ivan"`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			token, err := a.stateManager.RequireToken()
			if err != nil {
				return err
			}
			a.apiClient.SetToken(token)

			state, err := a.stateManager.Load()
			if err != nil || len(state.Salt) == 0 {
				return fmt.Errorf("соль не найдена — зарегистрируйтесь заново или восстановите состояние")
			}

			password, err := readMasterPassword()
			if err != nil {
				return err
			}

			key := crypto.DeriveKey(password, state.Salt, state.ArgonParams)

			plaintext, cType, err := buildPlainText(credType, username, password, text, filePath, cardNumber, cardExpiry, cardCVV, cardHolder)
			if err != nil {
				return err
			}

			encrypted, err := crypto.Encrypt(plaintext, key)
			if err != nil {
				return fmt.Errorf("шифрование: %w", err)
			}

			cred, err := a.apiClient.CreateCredential(cmd.Context(), cType, name, metadata, encrypted)
			if err != nil {
				return fmt.Errorf("создание записи: %w", err)
			}

			fmt.Printf("Запись создана: ID=%s, имя=%s, тип=%s\n", cred.ID, cred.Name, cred.Type)
			return nil
		},
	}

	cmd.Flags().StringVar(&credType, "type", "", "тип записи: login_password, text, binary, bank_card (обязательно)")
	cmd.Flags().StringVar(&name, "name", "", "название записи (обязательно)")
	cmd.Flags().StringVar(&metadata, "meta", "", "дополнительные метаданные")
	cmd.Flags().StringVar(&username, "username", "", "имя пользователя (для login_password)")
	cmd.Flags().StringVar(&text, "text", "", "текстовое содержимое (для text)")
	cmd.Flags().StringVar(&filePath, "file", "", "путь к файлу (для binary)")
	cmd.Flags().StringVar(&cardNumber, "number", "", "номер карты (для bank_card)")
	cmd.Flags().StringVar(&cardExpiry, "expiry", "", "срок действия карты (для bank_card)")
	cmd.Flags().StringVar(&cardCVV, "cvv", "", "CVV-код (для bank_card)")
	cmd.Flags().StringVar(&cardHolder, "holder", "", "имя держателя карты (для bank_card)")

	_ = cmd.MarkFlagRequired("type")
	_ = cmd.MarkFlagRequired("name")

	return cmd
}

// buildPlainText формирует JSON-данные для шифрования в зависимости от типа.
func buildPlainText(
	credType, username, password, text, filePath, cardNumber, cardExpiry, cardCVV, cardHolder string,
) ([]byte, model.CredentialType, error) {
	switch model.CredentialType(credType) {
	case model.CredentialTypeLoginPassword:
		data := loginPasswordData{Username: username, Password: password}
		b, err := json.Marshal(data)
		return b, model.CredentialTypeLoginPassword, err

	case model.CredentialTypeText:
		if text == "" {
			return nil, "", fmt.Errorf("для типа text укажите --text")
		}
		data := textData{Text: text}
		b, err := json.Marshal(data)
		return b, model.CredentialTypeText, err

	case model.CredentialTypeBinary:
		if filePath == "" {
			return nil, "", fmt.Errorf("для типа binary укажите --file")
		}
		content, err := os.ReadFile(filePath)
		if err != nil {
			return nil, "", fmt.Errorf("чтение файла: %w", err)
		}
		data := binaryData{Filename: filePath, Content: content}
		b, err := json.Marshal(data)
		return b, model.CredentialTypeBinary, err

	case model.CredentialTypeBankCard:
		data := bankCardData{Number: cardNumber, Expiry: cardExpiry, CVV: cardCVV, Holder: cardHolder}
		b, err := json.Marshal(data)
		return b, model.CredentialTypeBankCard, err

	default:
		return nil, "", fmt.Errorf("неизвестный тип '%s', доступно: login_password, text, binary, bank_card", credType)
	}
}
