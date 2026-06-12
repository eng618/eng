# Technical Roadmap & Action Plan

This document serves as the authoritative roadmap for technical improvements, modernization, and technical debt reduction for the `eng` CLI application. It is based on a comprehensive senior-level code review and is designed to be executed incrementally.

## Status Tracking Mechanism

Use the following markers to track the status of each task:

- `[ ]` Not Started
- `[/]` In Progress
- `[!]` Blocked
- `[x]` Complete

---

## 1. Critical & High Priority (Immediate Fixes & Quick Wins)

These items represent immediate reliability issues and simple fixes that provide significant value.

### 1.1 Fix Unhandled Errors in File Operations

- **Status**: `[x]`
- **Task**: Wrap all `os.WriteFile` and `os.MkdirAll` calls in the `codemod` package with proper error handling (`if err != nil { return err }`).
- **Impact**: High (prevents silent failures during code generation)
- **Effort**: Low (Under 1 hour)
- **Location**: `cmd/codemod/native.go`, `cmd/codemod/web.go`

### 1.2 Remove `log.Fatal` from Library Packages

- **Status**: `[x]`
- **Task**: Remove all `log.Fatal()` and `os.Exit(1)` calls from the `internal/` directory. Return `error` types to the caller so the top-level command can handle fatal exits gracefully.
- **Impact**: Critical (improves reliability and testability)
- **Effort**: Low
- **Location**: `internal/utils/command.go`, `internal/utils/log/log.go`

### 1.3 Auto-Format Codebase

- **Status**: `[x]` Complete
- **Task**: Run `golangci-lint run --fix` or `gci` to auto-correct imports and style formatting issues.
- **Impact**: Low
- **Effort**: Low
- **Location**: Repository-wide

---

## 2. Medium Priority (Architecture & Performance)

These items address structural debt and optimize the CLI's performance for scale.

### 2.1 Refactor `internal/utils` into Domain Packages

- **Status**: `[x]`
- **Task**: Break apart the `internal/utils` grab-bag package into dedicated, single-responsibility domain packages (e.g., `internal/repo`, `internal/bitwarden`, `internal/config`, `internal/ui`).
- **Impact**: High (improves maintainability and separation of concerns)
- **Effort**: Medium
- **Location**: `internal/utils/*`

### 2.2 Parallelize Multi-Repository Operations

- **Status**: `[x]`
- **Task**: Introduce `golang.org/x/sync/errgroup` to perform `git fetch` and `git pull` operations concurrently across projects and repositories.
- **Impact**: High (drastically reduces latency for syncing large numbers of repos)
- **Effort**: Medium
- **Location**: `cmd/project/sync.go`, `cmd/git/fetch_all.go`

### 2.3 Extract Business Logic from Cobra Commands

- **Status**: `[x]`
- **Task**: Move business logic out of Cobra's `Run`/`RunE` functions and into independent, testable service functions. The Cobra layer should only handle flag parsing, I/O routing, and service initialization.
- **Impact**: High (enables unit testing)
- **Effort**: High
- **Location**: `cmd/*`

### 2.4 Implement Context Propagation

- **Status**: `[x]`
- **Task**: Replace `exec.Command` with `exec.CommandContext` and pass a `context.Context` from the root command down to all network-bound or long-running operations. This allows users to cancel operations gracefully.
- **Impact**: Medium
- **Effort**: Medium
- **Location**: `internal/utils/repo/repo.go`, `internal/utils/bitwarden.go`

### 2.5 Replace `os/exec` with `go-git` where practical

- **Status**: `[x]`
- **Task**: Transition simple git system calls to use the native `go-git` library for better cross-platform compatibility and resilience.
- **Impact**: Medium
- **Effort**: Low
- **Location**: `internal/repo/*`, `internal/dotfiles/*`

---

## 3. Low Priority (Testing & Developer Experience)

These items ensure the long-term health and stability of the project.

### 3.1 Increase Test Coverage

- **Status**: `[x]`
- **Task**: Add table-driven unit tests for the newly extracted domain services (from task 2.3). Implement interfaces for external dependencies (like Git and FS) to allow mocking. Replace "no panic" pseudo-tests with actual assertions.
- **Impact**: High
- **Effort**: High
- **Location**: `cmd/*`, `internal/*`

### 3.2 Inject Configuration Dependencies

- **Status**: `[x]`
- **Task**: Replace internal, global `viper.GetString()` calls with explicit configuration structs passed to domain services.
- **Impact**: Medium (improves testability)
- **Effort**: Medium
- **Location**: Across the codebase

### 3.3 Replace Brittle Flag Lookups

- **Status**: `[x]`
- **Task**: Remove brittle `cmd.Parent().PersistentFlags()` lookups in subcommands. Bind flags to Viper or a configuration struct during `PreRun`.
- **Impact**: Low
- **Effort**: Low
- **Location**: `cmd/*`

### 3.4 Implement a TUI for Multi-Repo Syncing

- **Status**: `[x]`
- **Task**: Integrate a Text User Interface library like `bubbletea` or `pterm` to display rich parallel progress bars when fetching or syncing multiple repositories.
- **Impact**: Low (UX improvement)
- **Effort**: Medium
- **Location**: `cmd/project/sync.go`, `cmd/git/*`

## 4. Phase 4: Enhancing External Tooling & UX

These items focus on bringing the application's CLI interactions up to modern standards by introducing beautiful interactive terminal user interfaces (TUIs) and better parallel operation visualization.

### 4.1 Replace `AlecAivazis/survey` with `charmbracelet/huh`

- **Status**: `[x]`
- **Task**: Replace the outdated `survey` package with Charm's `huh` library for all user prompts. `huh` is built on Bubble Tea and offers superior theming, validation, and layout capabilities.
- **Impact**: High (modernizes the CLI feel, making configuration and prompts feel premium)
- **Effort**: Medium
- **Location**:
  - `cmd/system/*` (setup, proxy, update, kill_process, gpg_setup, etc.)
  - `cmd/files/*` (find_and_delete, find_non_movie_folders)
  - `cmd/project/*` (add, remove)
  - `cmd/config/*`
  - `cmd/dotfiles/*`

### 4.2 Expand `MultiSpinner` to File and System Commands

- **Status**: `[x]`
- **Task**: The new `ui.MultiSpinner` added during Phase 3 Git refactors is highly effective for visualizing concurrent or sequential long-running tasks. We will integrate it into other heavy commands.
- **Impact**: Medium (drastically improves the user experience during long wait times)
- **Effort**: Medium
- **Location**:
  - `cmd/files/find_and_delete.go`: Use a MultiSpinner for parallel file deletion (e.g., showing 4 worker threads deleting files concurrently).
  - `cmd/files/find_non_movie_folders.go`: Show concurrent directory scanning progress.
  - `cmd/system/update.go`: Use MultiSpinner for `apt-get`, `brew`, `asdf`, and cleanup steps instead of single blocking spinners.
  - `cmd/dotfiles/copy_changes.go`: Show concurrent diffs and copies.

## 5. Phase 5: Advanced TUI Enhancements & Wizards

Based on recent code exploration, these areas of the application can be enhanced with richer interactive terminal UIs using Bubble Tea or `huh`:

### 5.1 Interactive File Selection

- **Status**: `[x]`
- **Task**: Replace static printed lists of files with interactive `huh.MultiSelect` or Bubble Tea tables to allow users to visually check/uncheck files before deleting them.
- **Location**:
  - `cmd/files/find_and_delete.go`
  - `cmd/files/find_non_movie_folders.go`

### 5.2 Interactive Process Management

- **Status**: `[x]`
- **Task**: Upgrade the `kill` command to list running processes in an interactive, filterable table instead of relying on basic text prompts or manual PID entry.
- **Location**: `cmd/system/kill_process.go`

### 5.3 Setup Wizard Modernization

- **Status**: `[x]`
- **Task**: Convert the sequential system setup prompts into a beautiful, multi-page wizard using `huh.NewForm()` with clear steps and theming.
- **Location**: `cmd/system/setup.go`, `cmd/system/gpg_setup.go`

### 5.4 Multi-Spinner System Updates

- **Status**: `[x]`
- **Task**: Integrate `ui.MultiSpinner` to display concurrent progress for system updates, dotfiles syncing, and cleaning tasks.
- **Location**:
  - `cmd/system/update.go`
  - `cmd/dotfiles/copy_changes.go`
