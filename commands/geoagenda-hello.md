---
description: Call the geoagenda /hello endpoint with the stored credentials
---

Call the geoagenda /hello endpoint and report the result:

```bash
${CLAUDE_PLUGIN_ROOT}/bin/geoagenda hello --json
```

Parse the JSON. On success, report `message`. If the exit code is `4`
(unauthenticated), tell the user to run `/geoagenda-login` — do not attempt
to drive login yourself. For any other non-zero exit, surface the error
message verbatim.
