# Manage Repositories in Bulk

Goal: keep many git repositories (grouped into projects) up to date with a
few commands, and inspect failures without re-running anything.

## Prerequisites

```sh
eng config git-dev-path ~/Development
```

Use `--current` on any `git` bulk command to target the current directory
instead of the configured path; use `--dry-run` to preview.

## Daily sync

```sh
eng git sync-all          # fetch + pull --rebase everywhere
eng git fetch-all         # fetch only
eng git pull-all          # pull only
eng git push-all          # push repos with unpushed commits
eng git push-all --force --yes   # force-with-lease (confirms unless --yes)
eng git status-all        # working-tree overview table
```

Project-scoped equivalents (add `-p <project>` to narrow any of them):

```sh
eng project fetch
eng project pull
eng project sync
```

## Work interactively

```sh
eng dashboard
```

Actions apply to the focused repo (right pane) or the whole project (left
pane). Multi-repo runs show a fixed-size progress modal with per-repo rows
(`○` queued, spinner running, `✓` done, `−` skipped, `✗` error).

## When something fails

Bulk commands print a summary and save full detail to a session log:

```sh
eng logs show --tail 50   # last 50 lines of the latest run
```

Project not configured yet?

```sh
eng project add -p MyProject
eng project setup
```

## See also

- [Session logs](session-logs.md)
- [Command reference: git](../reference/cli/eng_git.md), [project](../reference/cli/eng_project.md)
- [Getting started](../tutorials/getting-started.md)
