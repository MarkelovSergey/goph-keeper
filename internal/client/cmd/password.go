package cmd

import (
	"fmt"
	"os"

	"golang.org/x/term"
)

// readMasterPassword запрашивает мастер-пароль у пользователя.
// Ввод скрыт — символы не отображаются в терминале.
// Переменная позволяет подменять реализацию в тестах.
var readMasterPassword = func() (string, error) {
	fmt.Print("Мастер-пароль: ")
	pass, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println()
	if err != nil {
		return "", fmt.Errorf("чтение пароля: %w", err)
	}
	return string(pass), nil
}
