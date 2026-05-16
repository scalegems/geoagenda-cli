---
description: Show which geoagenda user is currently logged in
---

Show the stored geoagenda identity:

```bash
${CLAUDE_PLUGIN_ROOT}/bin/geoagenda whoami --json
```

If exit code is `4`, the user is not logged in — suggest `/geoagenda-login`.
Otherwise report the email and user_id from the JSON output.
