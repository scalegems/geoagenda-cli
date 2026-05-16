package cmd

import (
	"geoagenda-cli/internal/store"

	"github.com/spf13/cobra"
)

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Remove stored credentials",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := store.Clear(); err != nil {
			fail(err.Error(), ExitGeneric)
		}
		emit("logged out", map[string]any{"ok": true})
		return nil
	},
}
