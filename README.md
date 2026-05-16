# geoagenda-cli

Command-line client for the geoagenda Convex backend. Authenticates with WorkOS
via OAuth 2.0 (PKCE + loopback redirect) and stores credentials in the OS
keychain.

## Install

### As a Claude Code plugin (recommended for agent use)

```
/plugin install scalegems/geoagenda-cli
```

The plugin's `Setup` hook fetches the matching binary from the latest GitHub
Release and installs it under the plugin's `bin/` directory. The plugin also
ships a skill that teaches the agent how to drive the CLI, plus the slash
commands `/geoagenda-login`, `/geoagenda-hello`, and `/geoagenda-whoami`.

### From a release archive

Download the archive for your OS/arch from the [Releases page](https://github.com/scalegems/geoagenda-cli/releases),
extract it, and put `geoagenda` somewhere on your `$PATH`.

### From source

Requires Go 1.22+.

```sh
go install github.com/scalegems/geoagenda-cli@latest
# or, from a clone:
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

## Release process

[`.github/workflows/release.yml`](.github/workflows/release.yml) runs GoReleaser
to cross-compile for `darwin/{amd64,arm64}`, `linux/{amd64,arm64}`, and
`windows/amd64`, attaches archives + a `checksums.txt`, and publishes a
GitHub Release. Two ways to trigger it:

**Manual (recommended):** GitHub UI → Actions → "release" → "Run workflow",
pick `patch` / `minor` / `major`, leave `version` blank. The workflow computes
the next semver from the latest existing tag, creates the tag, pushes it, and
runs GoReleaser. Or from a terminal:

```sh
gh workflow run release.yml -f bump=minor
# explicit version overrides bump
gh workflow run release.yml -f version=v0.2.0-rc.1
gh run watch
```

**Tag push (manual semver):**

```sh
git tag v0.1.0
git push origin v0.1.0
```

**Dry-run locally** (no tag created, no upload):

```sh
goreleaser release --snapshot --clean
```
