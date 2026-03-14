package cmd

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/spf13/cobra"

	"github.com/MarkelovSergey/goph-keeper/internal/client/api"
)

func newDeleteCmd() *cobra.Command {
	var id string

	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Удалить запись по ID",
		RunE: func(cmd *cobra.Command, _ []string) error {
			token, err := stateManager.RequireToken()
			if err != nil {
				return err
			}
			apiClient.SetToken(token)

			credID, err := uuid.Parse(id)
			if err != nil {
				return fmt.Errorf("неверный формат ID: %w", err)
			}

			err = apiClient.DeleteCredential(cmd.Context(), credID)
			if errors.Is(err, api.ErrNotFound) {
				return fmt.Errorf("запись не найдена")
			}
			if err != nil {
				return fmt.Errorf("удаление записи: %w", err)
			}

			fmt.Printf("Запись %s удалена.\n", id)
			return nil
		},
	}

	cmd.Flags().StringVar(&id, "id", "", "UUID записи (обязательно)")
	_ = cmd.MarkFlagRequired("id")

	return cmd
}
