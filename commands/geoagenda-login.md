---
description: Log in to geoagenda via WorkOS OAuth (opens a browser)
---

Run the geoagenda login flow for the user:

```bash
${CLAUDE_PLUGIN_ROOT}/bin/geoagenda login
```

Tell the user a browser is about to open and they'll need to complete the
WorkOS sign-in. Wait for the command to return, then report the resulting
identity (or the error). If the command exits non-zero, surface the error
message verbatim — don't retry login automatically.
