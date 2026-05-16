package cmd

import (
	"geoagenda-cli/internal/store"

	"github.com/spf13/cobra"
)

var whoamiCmd = &cobra.Command{
	Use:   "whoami",
	Short: "Show the currently authenticated user",
	RunE: func(cmd *cobra.Command, args []string) error {
		sess, err := store.Load()
		if err != nil {
			fail("not authenticated — run `geoagenda login`", ExitUnauth)
		}
		emit(
			"logged in as "+sess.Email+" ("+sess.UserID+")",
			map[string]any{
				"user_id":         sess.UserID,
				"email":           sess.Email,
				"organization_id": sess.OrganizationID,
				"issuer":          sess.Issuer,
			},
		)
		return nil
	},
}
