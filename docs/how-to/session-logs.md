# Inspect Session Logs

Goal: see full detail behind a clean terminal summary without re-running a
verbose command.

## Background

Verbose commands (`git sync-all/fetch-all/pull-all/push-all`,
`project fetch/pull/sync`, `system update`) print a summary and capture
everything to a timestamped file (latest 20 runs kept; `ENG_LOG_DIR`
overrides the location). Each run ends with:

```text
→ Full log: <path> (view with `eng logs show`)
```

## View logs

```sh
eng logs                  # list recent sessions, newest first
eng logs show             # show the latest run
eng logs show git-sync-all        # match by prefix
eng logs show --tail 50           # last 50 lines only
eng logs show system-update -f    # stream new lines until Ctrl+C
eng logs clean            # delete all session logs
```

## Tips

- Prefer `eng logs show` over `-v` re-runs when diagnosing a failure.
- In scripts/CI, logs are still captured; the terminal stays quiet.
- File logging is skipped in tests and never blocks a command on I/O errors.

## See also

- [Manage repositories](manage-repositories.md)
- [Command reference: logs](../reference/cli/eng_logs.md)
