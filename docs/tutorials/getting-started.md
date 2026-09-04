# Getting Started with `eng`

This tutorial takes you from zero to a working `eng` setup: install the CLI,
run first-time configuration, register your first project, and open the
interactive dashboard. Expect about ten minutes.

## 1. Install

Pick one method (macOS or Linux):

```sh
# One-line script (no Go or Homebrew required)
curl -sSfL https://raw.githubusercontent.com/eng618/eng/main/install.sh | sh

# …or Homebrew
brew tap eng618/eng
brew install eng
```

Verify the install:

```sh
eng version
eng doctor
```

`eng doctor` checks required tools (git, brew, bash) and your workspace paths.
It exits non-zero when something required is missing, so you can gate scripts on it.

## 2. First run: the setup wizard

The first time `eng` creates `~/.eng.yaml`, it offers a quick setup wizard
(email, dev folder, defaults) when run interactively:

```sh
eng git status-all
# → "First run detected. Run quick setup now?"
```

Accept and the wizard configures the essentials. Decline (or pipe/CI usage)
and nothing blocks you — later errors always suggest the exact fix, e.g.
`eng config edit --interactive`.

Prefer manual control? Skip the wizard and configure directly:

```sh
eng config git-dev-path ~/Development
eng config edit --interactive   # full TUI editor: email, dotfiles, GitLab, …
```

## 3. Register your first project

A _project_ is a named group of repositories living under your dev folder:

```sh
eng project add -p MyProject     # add a repo (prompts for URL)
eng project setup                # clone anything missing
eng project list                 # verify
```

## 4. Sync everything and open mission control

```sh
eng git sync-all                 # fetch + pull --rebase across all repos
eng dashboard                    # interactive project & git dashboard
```

In the dashboard: `Tab` switches panes, `1`/`2` switch tabs, `/` filters,
`f`/`p`/`s` fetch/pull/sync, `r` refreshes, `a` adds repos, `?` shows all
shortcuts, `q` quits.

Verbose commands print a clean summary and save full detail to a session log:

```sh
eng logs show                    # inspect the latest run
```

## Where next

- New machine? Follow [Set up a new machine](../how-to/new-machine-setup.md).
- Daily repo workflows: [Manage repositories](../how-to/manage-repositories.md).
- Dotfiles + secrets: [Back up and restore dotfiles secrets](../how-to/dotfiles-secrets.md).
- Full command details: [Command reference](../reference/commands.md), [per-command pages](../reference/cli/eng.md).
- How it all fits together: [Architecture](../explanation/architecture.md).
