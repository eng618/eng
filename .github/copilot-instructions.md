# Copilot Instructions for eng

These guidelines make AI agents productive in this Go CLI codebase by codifying architecture, workflows, and house conventions. Keep it concise and concrete.

## Big Picture

- CLI framework: Cobra entry at [cmd/root.go](cmd/root.go); `main()` delegates to `cmd.Execute()` in [main.go](main.go).
- Module layout: Subcommands under [cmd/](cmd) mirror features (git, dotfiles, system, codemod, ts, version, config, gitlab).
- Configuration: Viper-backed YAML at `$HOME/.eng.yaml` created/migrated on startup; see [utils/config/migration.go](utils/config/migration.go).
- Logging: Local colorized logger in [utils/log/log.go](utils/log/log.go); use it for all output and for `io.Writer`s.
- External tooling: Uses `glab`, `bw` (Bitwarden CLI), Homebrew, git-chglog, goreleaser where relevant (see [README.md](README.md)).

## Development Workflow

- Build: `task build` (outputs `./eng`) or `go build`; install with `task install` or `go install`.
- Lint/format: `task lint` (golangci-lint). Use `task format` / `task format-check` before commits.
- Test: `task test` runs `go test ./... -cover -race`.
- Validate: `task validate` runs format + lint + test.
- Releases/changelog: `task release`, `task changelog` (see [Taskfile.yaml](Taskfile.yaml)).
- Version ldflags: Build injects `Version`, `Commit`, `Date` for [cmd/version/version.go](cmd/version/version.go) using Taskfile `GO_BUILD_FLAGS`.

## Conventions

- Logging: Prefer `log.Start/Success/Info/Debug/Warn/Error` and `log.Verbose(v, ...)` for noisy output.
- Command output piping: For spawned processes, always wire `Stdout = log.Writer()` and `Stderr = log.ErrorWriter()`.
- Verbose flag: Respect `--verbose` via `utils.IsVerbose(cmd)` or viper key `verbose`.
- Errors: Return errors from helpers; let Cobra’s `CheckErr` or caller handle. Use `log.Fatal` only for unrecoverable process-level failures.
- Modules: Keep feature code in `cmd/<feature>` and shared helpers in `utils/` or feature-scoped packages under `utils/`.
- Tests: Use `testify` and prefer table-driven tests. When asserting logs, swap writers via `log.SetWriters` and `log.ResetWriters`.

## Patterns & Examples

- Spawn a child process with proper signal/IO handling:
  - Use `utils.StartChildProcess(exec.Command("tool", "args"...))` from [utils/command.go](utils/command.go).
- Bitwarden integration:
  - Use `utils.EnsureBitwardenSession()` and helpers in [utils/bitwarden.go](utils/bitwarden.go) to read/store secrets; do not parse `bw` JSON manually.
- Config keys stability:
  - Add migrations in [utils/config/migration.go](utils/config/migration.go) when renaming keys; call `viper.WriteConfig()` after changes.
- Multi-repo git ops:
  - Use commands in [cmd/git](cmd/git) that traverse the configured dev path (`eng config git-dev-path`), honoring `--current` and `--dry-run` where applicable.

## Integration Points

- Dotfiles: Managed as a bare repo via go-git; top-level commands in [cmd/dotfiles](cmd/dotfiles). Config keys: `dotfiles-repo-url`, `dotfiles-branch`, `dotfiles-bare-repo-path`.
- GitLab MR rules: See [cmd/gitlab](cmd/gitlab) for `mr-rules init/apply`; relies on `glab` and optional Bitwarden token retrieval.
- Codemods: [cmd/codemod](cmd/codemod) includes `lint-setup`, `prettier`, and `copilot` templates under [cmd/codemod/\*.tmpl](cmd/codemod).
- TypeScript helpers: Lightweight `ts up/down` commands in [cmd/ts](cmd/ts).

## Commit Messages

- Use Conventional Commits (feat, fix, chore, docs, refactor, test, style). Scope is encouraged, e.g., `feat(git): ...`.

## Keep In Sync

- Update [README.md](README.md) when adding commands, flags, or workflows.
- Prefer adding `task` targets for repeatable workflows and reference them here and in the README.

## Quick Checklist for New Code

- Wire logs via `utils/log`; respect `--verbose`.
- Add tests with `testify`; run `task validate`.
- Update config migrations if any keys change.
- Document new commands in [README.md](README.md) and add `task` targets if appropriate.
