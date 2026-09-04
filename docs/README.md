# `eng` Documentation

Documentation follows the [Diátaxis](https://diataxis.fr/) framework: each page
has one job — teach, guide, inform, or explain.

## Tutorials — learn by doing

Start here if `eng` is new to you. Tutorials are step-by-step lessons that
hold your hand to a working result.

- [Getting started](tutorials/getting-started.md) — install, first-run wizard,
  first project, dashboard tour (~10 minutes).

## How-to guides — solve a task

Goal-oriented recipes. They assume a working setup and get straight to the steps.

- [Set up a new machine](how-to/new-machine-setup.md) — full `system setup` walkthrough.
- [Manage repositories in bulk](how-to/manage-repositories.md) — git/project sync workflows.
- [Back up and restore dotfiles secrets](how-to/dotfiles-secrets.md) — `bws`-backed env files.
- [Inspect session logs](how-to/session-logs.md) — `eng logs` for past runs.
- [Release a new version](how-to/release.md) — Release Please + GoReleaser + Homebrew tap.
- [Test the codebase](how-to/test.md) — BET methodology, mocks, coverage.

## Reference — look things up

Dry, complete, exhaustive. No narrative, just facts.

- [Command reference](reference/commands.md) — hand-written reference for every command.
- [Per-command pages](reference/cli/eng.md) — generated from `--help` (`task docs` to refresh).
- [`gitlab-rules.json` example](reference/gitlab-rules.example.json) — MR rules schema example.

## Explanation — understand the system

Background, design rationale, and context.

- [Architecture](explanation/architecture.md) — modules, principles, config, dependencies.
- [Telemetry](explanation/telemetry.md) — privacy model, event schema, OpenPanel dashboards.
- [Roadmap](ROADMAP.md) — completed UI/UX roadmap and dropped items (project history).
