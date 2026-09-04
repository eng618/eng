# eng Running Plan

A living list of where this project stands and what's next. Items move from
**Active** to **Completed** (never deleted); **Dropped** records decisions
with rationale so they aren't relitigated.

Status key: `[ ]` not started · `[/]` in progress · `[x]` complete · `[!]` dropped

---

## Active next

- [x] **Dotfiles target-repo resolution** — `copy-changes` only found its
      destination via dev-folder heuristics; repos outside `devPath` broke it.
      Shipped: `eng config dotfiles-target-repo-path [path]`,
      `copy-changes --repo`, no devPath requirement with explicit targets,
      fail-loud validation, interactive locate-and-persist fallback.
- [ ] **A4 docs content gaps** — `reference/commands.md` drift (Dashboard TOC
      entry, `compose clean`, codemod `native`/`web`, full config rows,
      `immich`/`doctor` sections; explain `eng immich` vs `eng system immich`);
      second tutorial.
- [ ] **Deferred file splits** — `cmd/system/gpg_setup.go`,
      `cmd/system/ssh_setup.go`, `internal/immich/client.go` (large but cohesive;
      split when they next change).
- [ ] **Progress-display interface** — spinners/pagers/terminal-width are still
      imported directly by 7 internal packages (`project`, `cleanup`,
      `dotfiles`, `secrets`, `immich`, `runlog`); prompts are already
      hook-injected. A `cmd/`-boundary progress interface is the remaining step
      if the coupling ever bites (untested interactive paths, reuse outside eng).
- [ ] **`cmdutil → config` wrinkle** — `IsVerbose` falls back to viper,
      keeping the helper from being a true leaf. Fix by pushing the fallback to
      callers if it ever matters.

---

## Completed

### UI/UX foundations (Phases 6–9)

- [x] **6.1 Global design system** — `internal/ui/theme` brand tokens,
      typography, borders (`gv-tech`).
- [x] **6.2 Rich command output** — all production output flows through
      unified `log.Out`/`log.Err` writers; machine stdout stays clean (`--json`
      prints only JSON, `proxy export` silences chatter for `eval`).
- [x] **7.1 Actionable errors** — styled error boxes with next-step
      suggestions; `eng doctor` exits non-zero for CI gating.
- [x] **7.2 Progressive disclosure** — session log files (`internal/runlog`,
      20-run rotation) + `eng logs list/show/clean`; verbose commands print
      summaries with a `Full log:` pointer.
- [x] **8.1 Project & git dashboard** — Bubble Tea mission control with
      fixed-size Docker-style action modal, compact/responsive layouts.
- [x] **8.2 Interactive config editor** — `eng config edit --interactive`
      huh TUI.
- [x] **9.1 First-run wizard** — one-time guarded setup offer
      (`config.ShouldAutoOnboard`: true first run, TTY, sane `TERM`, skips
      config/meta/help/completion, `ENG_NO_ONBOARDING` escape).
- [x] **9.3 Dynamic auto-completion** — project names, compose stacks,
      proxy titles, shred methods, log names, `up --env`.

### Docs (Diátaxis)

- [x] Restructure (`tutorials/`, `how-to/`, `reference/`, `explanation/`,
      index) with refreshed architecture doc.
- [x] Automation: `gendocs --check/--prune`, `task docs:check` in
      `validate`, markdownlint + prettier + link-check tasks, docs CI workflow.

### Separation of concerns (B1–B4)

- [x] **B1 layering repairs** — `internal/version` leaf (ldflags updated);
      dotfiles setup steps injected via `cmd/system/dotfiles_hooks.go` instead
      of `internal → cmd` imports; immich tree built once by
      `internal/immich.NewCommand` with independent flag state per path.
- [x] **B2 UI decoupling** — dashboard DTOs + provider (`ui` no longer
      imports `config`/`containers`); config/dotfiles prompts via hook vars;
      URL helpers moved to `internal/repo`; tag-clobber prompt moved to
      `internal/project`.
- [x] **B3 dedup** — single `ui.Truncate`; single `cmdutil.CompletePrefix`.
- [x] **B4 splits** — dashboard `update`/`view` → `actions`/`status`/
      `commands`/`modal`/`table`; `proxy.go` → view/commands;
      `config/proxy.go` → env/manage/prompt.
- [x] **Dependency rules** recorded in `explanation/architecture.md`.

---

## Dropped

- [!] **9.2 Glamour help rendering** — help is already custom-styled
  (colored headers, aligned columns, populated `EXAMPLES`); Glamour would
  add a heavy dependency for marginal gain.
