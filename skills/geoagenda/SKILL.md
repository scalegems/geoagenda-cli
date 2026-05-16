---
name: geoagenda
description: Use this skill whenever the user asks to interact with the geoagenda backend from the command line — checking who they are logged in as, calling backend endpoints, or fetching a fresh access token. Authenticates via WorkOS OAuth (PKCE + browser) and stores credentials in the OS keychain.
---

# geoagenda CLI

A Go CLI that authenticates against WorkOS and talks to the geoagenda Convex
backend. The binary lives at `${CLAUDE_PLUGIN_ROOT}/bin/geoagenda` and is
already on the Bash tool's `$PATH` while this plugin is installed.

## When to use this skill

Reach for the CLI when the user wants to:

- Check whether they're authenticated against geoagenda (`whoami`)
- Call the `/hello` endpoint to verify the auth pipeline end-to-end
- Get a fresh access token for use in other tooling
- Log in or log out

Do **not** reach for it for tasks that don't involve the geoagenda backend.

## Commands

| Command | What it does | Interactive? |
|---|---|---|
| `geoagenda login` | Opens a browser, runs the WorkOS OAuth PKCE flow, stores refresh token in the OS keychain. | **Yes** — requires a human + a browser. |
| `geoagenda whoami` | Prints the stored identity. | No |
| `geoagenda hello` | Calls the Convex `/hello` HTTP route with a fresh access token. | No |
| `geoagenda logout` | Deletes the stored session. | No |

Every command accepts `--json` for structured output. Always pass `--json`
when you intend to parse the result.

## Config

The CLI reads two env vars (or accepts equivalent flags):

- `WORKOS_CLIENT_ID` — same as `VITE_WORKOS_CLIENT_ID` from the geoagenda app.
- `CONVEX_URL` — the deployment URL. Either `.convex.cloud` or `.convex.site`
  is accepted; the CLI normalizes internally.

If either is missing, the CLI exits with code 2 and a clear message naming
the missing variable.

## Exit codes

| Code | Meaning | What to do |
|---|---|---|
| `0` | Success | — |
| `1` | Generic failure | Read the error, decide whether to retry. |
| `2` | Bad usage / missing config | Ask the user for the missing env var; don't guess. |
| `4` | Unauthenticated | Tell the user to run `geoagenda login` themselves. Do **not** attempt to drive `login` yourself — it opens a browser. |
| `5` | Forbidden | Stop and report; this is not a retryable error. |

## Output shapes (with `--json`)

```jsonc
// geoagenda whoami --json
{
  "user_id": "user_01HABC...",
  "email": "alice@example.com",
  "organization_id": "org_01H...",
  "issuer": "https://api.workos.com/user_management"
}

// geoagenda hello --json
{ "message": "hello" }

// any error path
{ "error": "missing --client-id (or WORKOS_CLIENT_ID env var)" }
```

## Critical rules

1. **Never run `geoagenda login` non-interactively.** It opens a browser and
   waits for a callback on `127.0.0.1:8765`. If the user is unauthenticated,
   instruct them to run `/geoagenda-login` (or `geoagenda login` in a
   terminal) themselves, then continue once they confirm.
2. **Always pass `--json` when parsing.** The non-JSON output is for humans
   and may change.
3. **Respect exit codes.** Exit `4` means re-running won't help — only login
   will. Don't loop.
4. **Don't print secrets.** The CLI never echoes tokens; don't ask it to,
   and don't try to read them from the keychain directly.

## Convenience slash commands

The plugin also installs two slash commands the human can invoke:

- `/geoagenda-login` — runs `geoagenda login` in their terminal
- `/geoagenda-hello` — runs `geoagenda hello --json` and reports the result
