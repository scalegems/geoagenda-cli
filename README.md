# geoagenda-cli

Command-line client for the geoagenda Convex backend. Authenticates with WorkOS
via OAuth 2.0 (PKCE + loopback redirect) and stores credentials in the OS
keychain.

## Build

Requires Go 1.22+.

```sh
cd cli
go mod tidy
go build -o geoagenda .
```

## Configure

Production defaults are baked in — `geoagenda login` works with no config.

Override per-shell to hit a different environment:

```sh
export WORKOS_CLIENT_ID="client_..."                   # same as VITE_WORKOS_CLIENT_ID
export CONVEX_URL="https://<deployment>.convex.cloud"  # .cloud or .site — the CLI handles both
```

Or per-invocation via flags: `--client-id`, `--convex-url`, `--issuer`, `--port`.

In your WorkOS dashboard, `http://127.0.0.1:8765/callback` must be in the
allowed redirect URIs for this client (or pass `--port` to use a different port).

## Commands

```sh
geoagenda login           # browser-based OAuth, stores refresh token in keychain
geoagenda whoami          # print the stored identity
geoagenda hello           # call Convex /hello with a fresh access token
geoagenda logout          # wipe stored credentials
```

Add `--json` to any command for machine-readable output.

## Exit codes

| Code | Meaning |
|------|---------|
| `0`  | success |
| `1`  | generic failure |
| `2`  | bad usage / missing config |
| `4`  | unauthenticated — run `geoagenda login` |
| `5`  | forbidden |

## Storage

Refresh token + identity metadata live in the OS keychain under service
`geoagenda-cli`, account `default`. Access tokens are short-lived and held in
memory only.

- macOS: Keychain
- Windows: Credential Manager (DPAPI)
- Linux: Secret Service (GNOME Keyring, KWallet)
