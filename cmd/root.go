package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

const (
	ExitOK        = 0
	ExitGeneric   = 1
	ExitUsage     = 2
	ExitUnauth    = 4
	ExitForbidden = 5
)

const (
	defaultIssuer    = "https://api.workos.com/user_management"
	defaultClientID  = "client_01K935TA6B1SD1VAAHDXTXJ9BZ"
	defaultConvexURL = "https://dapper-puffin-175.convex.cloud"
)

var (
	jsonOutput bool
	clientID   string
	convexURL  string
	issuer     string
	port       int
)

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

var rootCmd = &cobra.Command{
	Use:           "geoagenda",
	Short:         "geoagenda CLI",
	Long:          "Command-line client for the geoagenda Convex backend.",
	SilenceUsage:  true,
	SilenceErrors: true,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fail(err.Error(), ExitGeneric)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, "Output machine-readable JSON")
	rootCmd.PersistentFlags().StringVar(&clientID, "client-id", envOr("WORKOS_CLIENT_ID", defaultClientID), "WorkOS client ID")
	rootCmd.PersistentFlags().StringVar(&convexURL, "convex-url", envOr("CONVEX_URL", defaultConvexURL), "Convex deployment URL (.cloud or .site)")
	rootCmd.PersistentFlags().StringVar(&issuer, "issuer", defaultIssuer, "WorkOS User Management base URL")
	rootCmd.PersistentFlags().IntVar(&port, "port", 8765, "Local port for the OAuth callback listener")

	rootCmd.AddCommand(loginCmd, logoutCmd, whoamiCmd, helloCmd)
}

func emit(human string, data any) {
	if jsonOutput {
		_ = json.NewEncoder(os.Stdout).Encode(data)
		return
	}
	fmt.Println(human)
}

func fail(msg string, code int) {
	if jsonOutput {
		_ = json.NewEncoder(os.Stderr).Encode(map[string]any{"error": msg})
	} else {
		fmt.Fprintln(os.Stderr, "error:", msg)
	}
	os.Exit(code)
}

func requireClientID() {
	if clientID == "" {
		fail("missing --client-id (or WORKOS_CLIENT_ID env var)", ExitUsage)
	}
}

func requireConvexURL() {
	if convexURL == "" {
		fail("missing --convex-url (or CONVEX_URL env var)", ExitUsage)
	}
}

// convexHTTPBase returns the host that serves Convex HTTP actions. Convex
// HTTP routes live at <deployment>.convex.site, while clients usually have
// the .convex.cloud URL configured. Translate transparently.
func convexHTTPBase(u string) string {
	u = strings.TrimRight(u, "/")
	if strings.HasSuffix(u, ".convex.cloud") {
		return strings.TrimSuffix(u, ".convex.cloud") + ".convex.site"
	}
	return u
}
