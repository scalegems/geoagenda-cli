package cmd

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"

	"geoagenda-cli/internal/auth"
	"geoagenda-cli/internal/store"

	"github.com/spf13/cobra"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with WorkOS and store credentials locally",
	RunE: func(cmd *cobra.Command, args []string) error {
		requireClientID()

		ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
		defer stop()

		sess, err := auth.Login(ctx, auth.LoginOptions{
			ClientID: clientID,
			Issuer:   issuer,
			Port:     port,
		})
		if err != nil {
			fail(err.Error(), ExitGeneric)
		}

		if err := store.Save(sess); err != nil {
			fail(fmt.Sprintf("failed to persist session: %v", err), ExitGeneric)
		}

		emit(
			fmt.Sprintf("logged in as %s (%s)", sess.Email, sess.UserID),
			map[string]any{
				"user_id":         sess.UserID,
				"email":           sess.Email,
				"organization_id": sess.OrganizationID,
			},
		)
		return nil
	},
}
