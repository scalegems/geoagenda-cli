---
name: geoagenda
description: Use this skill whenever the user wants to interact with the geoagenda backend — checking who they're logged in as, calling backend endpoints, fetching access tokens, or logging in/out. Works in two modes — drives a local geoagenda CLI binary when available (Claude Code with the geoagenda-cli plugin), or guides the user through running it themselves when not (claude.ai, Claude Cowork, Managed Agents).
---

# geoagenda

A Go CLI that authenticates against WorkOS (OAuth 2.0 PKCE + browser) and talks
to the geoagenda Convex backend.

## Pick a mode before doing anything

This skill runs on more than one Claude surface. Figure out which mode you're
in *first*; the wrong mode produces confusing errors or hallucinated output.

- **Active mode** — you can execute shell commands **AND** the file
  `${CLAUDE_PLUGIN_ROOT}/bin/geoagenda` exists. This is Claude Code with the
  geoagenda-cli plugin installed. Drive the CLI directly.
- **Advisory mode** — you can't execute shell commands, or the binary isn't
  present. This is claude.ai, Claude Cowork without the plugin, Managed
  Agents, or any environment where `${CLAUDE_PLUGIN_ROOT}` doesn't resolve.
  Don't try to invoke anything — guide the user instead.

If unsure, default to **advisory**. Running a binary that doesn't exist is
worse than asking the user to confirm.

## When to use this skill

In either mode, reach for it when the user wants to:

- Check whether they're authenticated against geoagenda (`whoami`)
- Call the `/hello` endpoint to verify the auth pipeline end-to-end
- Get a fresh access token for other tooling
- Log in or log out

Don't reach for it for anything unrelated to the geoagenda backend.

---

## Active mode (Claude Code with plugin)

The binary lives at `${CLAUDE_PLUGIN_ROOT}/bin/geoagenda` — the plugin's
install hook put it there. Call it via the Bash tool.

### Commands

| Command | What it does | Interactive? |
|---|---|---|
| `geoagenda login` | Opens a browser, runs the WorkOS OAuth PKCE flow, stores refresh token in the OS keychain. | **Yes** — requires a human + a browser. |
| `geoagenda whoami` | Prints the stored identity. | No |
| `geoagenda hello` | Calls the Convex `/hello` HTTP route with a fresh access token. | No |
| `geoagenda logout` | Deletes the stored session. | No |

Always pass `--json` when you intend to parse output.

### Config

The CLI reads two env vars (or accepts equivalent flags):

- `WORKOS_CLIENT_ID` — same as `VITE_WORKOS_CLIENT_ID` from the geoagenda app.
- `CONVEX_URL` — the deployment URL. Either `.convex.cloud` or `.convex.site`
  is accepted; the CLI normalizes internally.

If either is missing, the CLI exits with code 2 and a message naming it.

### Exit codes

| Code | Meaning | What to do |
|---|---|---|
| `0` | Success | — |
| `1` | Generic failure | Read the error, decide whether to retry. |
| `2` | Bad usage / missing config | Ask the user for the missing env var; don't guess. |
| `4` | Unauthenticated | Tell the user to run `geoagenda login` themselves. Do **not** drive `login` — it opens a browser. |
| `5` | Forbidden | Stop and report; not retryable. |

### Output shapes (with `--json`)

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

---

## Advisory mode (claude.ai, Cowork without plugin, Managed Agents)

You can't run the binary here, but you can still help. When the user asks
for something the CLI would do:

1. **Don't invoke anything.** No Bash, no Python `subprocess`, no faked output.
2. **Direct them to install it:**
   - Claude Code: `/plugin install scalegems/geoagenda-cli`
   - Terminal: download the binary for their OS from
     <https://github.com/scalegems/geoagenda-cli/releases>
3. **Show them the exact command** they'd run, including `--json` and any
   flags relevant to the question.
4. **Offer to interpret the output** if they paste the JSON back.

Example phrasing:

> I can't reach geoagenda from this surface — it needs the `geoagenda` CLI
> running on your machine. Install it with `/plugin install
> scalegems/geoagenda-cli` in Claude Code, or grab the binary from
> <https://github.com/scalegems/geoagenda-cli/releases>. Then run
> `geoagenda whoami --json` and paste the result here.

In advisory mode, your job is to be the documentation the user wishes the
tool had — accurate commands, expected output shapes, and what each error
means.

---

## Critical rules (both modes)

1. **Never run `geoagenda login` non-interactively.** It opens a browser and
   waits for a callback on `127.0.0.1:8765`. If the user is unauthenticated,
   ask them to run `/geoagenda-login` (Claude Code) or `geoagenda login`
   (terminal) themselves; resume once they confirm.
2. **Always pass `--json` when parsing.** Non-JSON output is for humans and
   may change.
3. **Respect exit codes.** Exit `4` means re-running won't help — only
   login will. Don't loop.
4. **Don't print secrets.** The CLI never echoes tokens; don't ask it to,
   and don't try to read them from the keychain directly.

## Slash commands (Claude Code only)

The plugin installs three convenience slash commands the human can invoke
directly:

- `/geoagenda-login` — runs `geoagenda login` in their terminal
- `/geoagenda-hello` — runs `geoagenda hello --json` and reports the result
- `/geoagenda-whoami` — shows current login state
