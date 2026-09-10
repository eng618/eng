# AGENTS.md — eng CLI (Go, Cobra + Viper + Bubble Tea)

## Verify before push
- `task validate` = format + lint + test + `docs:check`. Always run before commits; CI (`go.yml`, `docs.yml`) runs the same gates.
- Focused checks: `go test ./internal/project/...`, `go test ./internal/repo/... -run TestName -v`.
- `task test` uses `-race` + CGO — needs `build-essential` on Linux.

## Generated docs are load-bearing
- Any change under `cmd/` (commands, flags, help text) requires `task docs` to regenerate `docs/reference/cli/`, then verify with `git status`.
- `task docs:check` runs in the pre-push hook (`git config core.hooksPath .githooks`) and CI. It fails on drift.
- Note: `.github/copilot-instructions.md` is stale (references `internal/utils/`, `cmd.Execute()`, `task release`) — trust this file and the Taskfile instead.

## Architecture
- Entry: `main.go` → `cmd.ExecuteContext` in `cmd/root.go`. Feature commands live in `cmd/<feature>/` as an exported `*Cmd` var registered in `root.go` with a `GroupID` (`devtools`/`envops`/`mgmt`/`meta`).
- Business logic lives in `internal/<domain>/` (`project`, `repo`, `config`, `ui`, …). There is no `internal/utils/` package.
- Boundaries are strict and tested:
  - `internal/project` takes pure-data `*Options` structs — never import Viper/config there; `cmd/` adapts config into DTOs.
  - Side effects go behind the `RepoClient` interface (`internal/project/repo_client.go`); tests use `MockRepoClient`.
  - Interactive prompts go through the `ConfirmPrompt` hook, wired to `ui.Confirm` in `cmd/` init — never import UI code from domain packages.
  - `internal/ui/dashboard` never reads config storage; `cmd/dashboard.go` injects projects + a reload provider.

## Style, organization, docs
- Format: `task format` (`gofmt -s`), lint-fix: `task lint-fix`. Import order enforced by `gci`: stdlib, default, then `prefix(github.com/eng618/eng)`. Line length 120 (`golines`); stricter `gofumpt` rules apply.
- Keep code organized by feature: new command → `cmd/<feature>/<feature>.go` + `init()` subcommand registration + `root.go` wiring. Shared logic → `internal/<domain>/` with a `doc.go` package overview.
- Document as you go: exported identifiers get godoc comments (`godot` lint); user-facing behavior goes in command `Long`/`Example` strings (which feed generated docs), not in code comments.
- Commits use Conventional Commits (`feat(git): …`) — `release-please` builds the changelog from them.

## Testing conventions
- `testify`, table-driven tests. For log assertions: `log.SetWriters(&buf, &buf)` + `defer log.ResetWriters()`.
- Tests touching spinners/progress must set `ui.DisableProgress = true`.
- `task build` drops a `./eng` binary in the repo root (gitignored) — don't commit it.
