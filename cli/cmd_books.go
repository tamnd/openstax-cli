package cli

import (
	"github.com/spf13/cobra"
)

func (a *App) booksCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "books",
		Short: "List all free OpenStax textbooks",
		RunE: func(cmd *cobra.Command, _ []string) error {
			limit := a.effectiveLimit(0)
			books, err := a.client.Books(cmd.Context(), limit)
			if err != nil {
				return mapFetchErr(err)
			}
			return a.renderOrEmpty(books, len(books))
		},
	}
	return cmd
}
