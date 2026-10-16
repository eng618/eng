# AGENTS.md — eng CLI (Go, Cobra + Viper + Bubble Tea)

## Verify before push

- `task validate` = format (Go) + `docs:format` (Markdown) + lint + test + `docs:check`. Always run before commits.
- `validate` does not cover Markdown lint/links; CI `docs.yml` also gates `task docs:lint`
  and `task docs:links` — run those when touching `*.md`.
- Markdown is formatted with Oxfmt via the `gv-oxfmt` wrapper (`@gv-tech/oxc-config` style),
  not Prettier: `task docs:format` to write, `task docs:lint` to verify. Same formatter as `eng fmt`.
- Focused checks: `go test ./internal/project/...`, `go test ./internal/repo/... -run TestName -v`.
- `task test` uses `CGO_ENABLED=1` + `-race` — needs `build-essential` on Linux.

## Generated docs are load-bearing

- Any change under `cmd/` (commands, flags, help text) requires `task docs` to regenerate
  `docs/reference/cli/`, then verify with `git status`. Never hand-edit generated pages.
- `task docs:check` runs in the pre-push hook (`git config core.hooksPath .githooks`) and CI. It fails on drift.

## Architecture

- Entry: `main.go` → `cmd.ExecuteContext` in `cmd/root.go`. Feature commands live in
  `cmd/<feature>/` as an exported `*Cmd` var registered in `root.go` with a `GroupID`
  (`devtools`/`envops`/`mgmt`/`meta`).
- Business logic lives in `internal/<domain>/` (`project`, `repo`, `config`, `ui`, …).
  There is no `internal/utils/` package; shared packages carry a `doc.go` overview.
- Boundaries are strict and tested:
  - `internal/project` takes pure-data `*Options` structs — never import Viper/config there; `cmd/` adapts config into DTOs.
  - Side effects go behind the `RepoClient` interface (`internal/project/repo_client.go`); tests use `MockRepoClient`.
  - Interactive prompts go through the `ConfirmPrompt` hook, wired to `ui.Confirm` in `cmd/` init — never import UI code from domain packages.
  - `internal/ui/dashboard` never reads config storage; `cmd/dashboard.go` injects projects + a reload provider.
- Config is Viper-backed YAML at `$HOME/.eng.yaml` with migrations in `internal/config`.

## Style, organization, docs

- Format: `task format` (`gofmt -s`), lint-fix: `task lint-fix`. Import order enforced by `gci`: stdlib, default, then `prefix(github.com/eng618/eng)`. Line length 120 (`golines`); stricter `gofumpt` rules and `godot` (exported identifiers need godoc comments) apply.
- Keep code organized by feature: new command → `cmd/<feature>/<feature>.go` + `init()` subcommand registration + `root.go` wiring. Shared logic → `internal/<domain>/` with a `doc.go` package overview.
- Document as you go: user-facing behavior goes in command `Long`/`Example` strings (which feed generated docs), not in code comments.
- Commits use Conventional Commits (`feat(git): …`) — releases are cut by `release-please` from them; there are no `task release`/`task changelog` targets.

## Testing conventions

- `testify`, table-driven tests. For log assertions: `log.SetWriters(&buf, &buf)` + `defer log.ResetWriters()`.
- Tests touching spinners/progress must set `ui.DisableProgress = true`.
- `task build` drops a `./eng` binary in the repo root (gitignored) — don't commit it.
