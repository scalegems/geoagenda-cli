package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"geoagenda-cli/internal/auth"
	"geoagenda-cli/internal/store"

	"github.com/spf13/cobra"
)

var helloCmd = &cobra.Command{
	Use:   "hello",
	Short: "Call the Convex /hello endpoint with a fresh access token",
	RunE: func(cmd *cobra.Command, args []string) error {
		requireClientID()
		requireConvexURL()

		ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
		defer stop()

		sess, err := store.Load()
		if err != nil {
			fail("not authenticated — run `geoagenda login`", ExitUnauth)
		}

		token, err := auth.AccessToken(ctx, sess, auth.RefreshOptions{
			ClientID: clientID,
			Issuer:   issuer,
		})
		if err != nil {
			fail(fmt.Sprintf("token refresh failed: %v", err), ExitUnauth)
		}
		// Refresh tokens rotate; persist any updates.
		if err := store.Save(sess); err != nil {
			fmt.Fprintln(cmd.ErrOrStderr(), "warning: could not persist rotated refresh token:", err)
		}

		url := convexHTTPBase(convexURL) + "/hello"
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			fail(err.Error(), ExitGeneric)
		}
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Accept", "application/json")

		client := &http.Client{Timeout: 15 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			fail(err.Error(), ExitGeneric)
		}
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)

		switch {
		case resp.StatusCode == http.StatusUnauthorized:
			fail(fmt.Sprintf("server returned 401: %s", body), ExitUnauth)
		case resp.StatusCode == http.StatusForbidden:
			fail(fmt.Sprintf("server returned 403: %s", body), ExitForbidden)
		case resp.StatusCode >= 400:
			fail(fmt.Sprintf("server returned %d: %s", resp.StatusCode, body), ExitGeneric)
		}

		var parsed any
		if err := json.Unmarshal(body, &parsed); err != nil {
			emit(string(body), map[string]any{"raw": string(body)})
			return nil
		}
		pretty, _ := json.MarshalIndent(parsed, "", "  ")
		emit(string(pretty), parsed)
		return nil
	},
}
