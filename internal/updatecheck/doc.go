/*
Package updatecheck notifies users about newer eng releases without a manual
`eng version` check, npm-style: a short stderr notice after the command runs.

It is designed to never slow down or break a command invocation:

  - The notice is served from a local cache file under the OS cache dir, so
    the common path performs one small file read and no network I/O.
  - Refreshing the cache happens in a fire-and-forget goroutine with a short
    timeout; failures are swallowed silently.
  - The check is skipped for dev builds, when ENG_NO_UPDATE_CHECK is set,
    for the `version` command itself (it does its own check), and in tests
    (via ui.DisableProgress, mirroring runlog).
*/
package updatecheck
