# eng CLI

```shell
                                          __ __
                                         |  \  \
  ______  _______   ______        _______| ▓▓\▓▓
 /      \|       \ /      \      /       \ ▓▓  \
|  ▓▓▓▓▓▓\ ▓▓▓▓▓▓▓\  ▓▓▓▓▓▓\    |  ▓▓▓▓▓▓▓ ▓▓ ▓▓
| ▓▓    ▓▓ ▓▓  | ▓▓ ▓▓  | ▓▓    | ▓▓     | ▓▓ ▓▓
| ▓▓▓▓▓▓▓▓ ▓▓  | ▓▓ ▓▓__| ▓▓    | ▓▓_____| ▓▓ ▓▓
 \▓▓     \ ▓▓  | ▓▓\▓▓    ▓▓     \▓▓     \ ▓▓ ▓▓
  \▓▓▓▓▓▓▓\▓▓   \▓▓_\▓▓▓▓▓▓▓      \▓▓▓▓▓▓▓\▓▓\▓▓
                  |  \__| ▓▓
                   \▓▓    ▓▓
                    \▓▓▓▓▓▓
```

[![Go](https://github.com/eng618/eng/actions/workflows/go.yml/badge.svg)](https://github.com/eng618/eng/actions/workflows/go.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/eng618/eng)](https://goreportcard.com/report/github.com/eng618/eng)
![GitHub Release](https://img.shields.io/github/v/release/eng618/eng)

A modern, modular CLI tool for developer automation, dotfiles management, system utilities, and project codemods. Built in Go, designed for extensibility and productivity.

---

## Features

- **Modular CLI**: Each command is a self-contained module (git, dotfiles, system, codemod, ts, version, config)
- **Git Repository Management**: Bulk operations across multiple git repositories with intelligent branch detection
- **Dotfiles Management**: Manage dotfiles via a bare git repo, with sync/fetch helpers
- **System Utilities**: MacOS/Linux helpers (kill port, find folders, proxy, update)
- **Codemod Automation**: Project codemods (e.g., lint setup for JS/TS projects)
- **TypeScript Helpers**: Up/down migration helpers for TypeScript projects
- **Config Management**: Centralized config via Viper
- **Versioning**: Shows build info, checks for updates, Homebrew auto-update
- **Taskfile-based Dev Workflow**: Build, test, lint, release, changelog automation
- **CI/CD**: GitHub Actions for lint, test, release, changelog

---

## Architecture

```mermaid
graph LR
    A[main.go] --> B[cmd/root.go]
    
    B --> C1[Git Commands]
    C1 --> D1[sync-all]
    C1 --> D2[fetch-all]
    C1 --> D3[pull-all]
    C1 --> D4[push-all]
    C1 --> D5[status-all]
    C1 --> D6[list]
    C1 --> D7[branch-all]
    C1 --> D8[stash-all]
    C1 --> D9[clean-all]
    
    B --> C2[Dotfiles]
    C2 --> E1[sync]
    C2 --> E2[fetch]
    
    B --> C3[System]
    C3 --> F1[kill-port]
    C3 --> F2[find-non-movie-folders]
    C3 --> F3[proxy]
    C3 --> F4[update]
    
    B --> C4[Other Commands]
    C4 --> G1[codemod]
    C4 --> G2[ts up/down]
    C4 --> G3[version]
    C4 --> G4[config]
    
    B --> C5[Utils]
    C5 --> H1[log]
    C5 --> H2[repo]
    C5 --> H3[config]
    C5 --> H4[files]
    C5 --> H5[command]
```

## Installation

### Homebrew (Recommended)

```sh
brew tap eng618/eng
brew install eng
```

### Go Install

```sh
go install github.com/eng618/eng@latest
```

### From Source

```sh
git clone https://github.com/eng618/eng.git
cd eng
go install .
```

---

## Usage

```sh
eng [command] [flags]
```

### Common Commands

- `eng git` — Git repository management across multiple repos
- `eng dotfiles` — Manage dotfiles (sync, fetch, info)
- `eng system` — System utilities (kill-port, find-non-movie-folders, update, proxy)
- `eng codemod` — Project codemods (e.g., lint-setup)
- `eng ts` — TypeScript helpers (up, down)
- `eng version` — Show version, check for updates
- `eng config` — Show or edit config

---

## Command Reference

### Git Repository Management

Manage multiple git repositories in your development folder with a comprehensive set of commands. All commands support the `--current` flag to operate on the current directory instead of the configured development path.

#### Setup

```sh
# Configure your development folder path
eng config git-dev-path /path/to/your/dev/folder
```

#### Repository Operations

- `eng git sync-all [--current] [--dry-run]` — Fetch and pull with rebase across all repositories
- `eng git fetch-all [--current] [--dry-run]` — Fetch latest changes from remote for all repositories  
- `eng git pull-all [--current] [--dry-run]` — Pull latest changes with rebase for all repositories
- `eng git push-all [--current] [--dry-run]` — Push local changes to remote for all repositories
- `eng git status-all [--current]` — Show git status for all repositories
- `eng git list [--current]` — List all git repositories found
- `eng git branch-all [--current]` — Show current branch for all repositories
- `eng git stash-all [--current] [--dry-run]` — Stash changes in all repositories
- `eng git clean-all [--current] [--dry-run]` — Clean untracked files in all repositories

#### Flags

- `--current` — Use current working directory instead of configured development path
- `--dry-run` — Show what would be done without making changes (where applicable)

### Dotfiles

- `eng dotfiles --info` — Show current dotfiles config
- `eng dotfiles sync` — Sync dotfiles repo
- `eng dotfiles fetch` — Fetch latest dotfiles

### System

- `eng system kill-port <port>` — Kill process on a port
- `eng system find-non-movie-folders [--dry-run]` — Find/delete non-movie folders
- `eng system update` — Update system packages
- `eng system proxy` — Manage proxy settings

### Codemod

- `eng codemod lint-setup` — Setup lint/format (eslint, prettier, husky, lint-staged) in JS/TS projects

### TypeScript

- `eng ts up` — Run up migration
- `eng ts down` — Run down migration

### Version

- `eng version` — Show version, build info, check for updates
- `eng version --update` — Auto-update via Homebrew (if installed that way)

### Config

- `eng config` — Show config
- `eng config edit` — Edit config
- `eng config git-dev-path` — Set development folder path for git commands

---

## Development Workflow

- **Build**: `task build` or `go build`
- **Install**: `task install` or `go install`
- **Lint**: `task lint` (uses golangci-lint)
- **Test**: `task test` (with coverage)
- **Validate**: `task validate` (lint + test)
- **Changelog**: `task changelog` (uses git-chglog)
- **Release**: `task release` (goreleaser + changelog)
- **Module Management**: `task tidy`, `task deps-upgrade`, etc.

See `Taskfile.yaml` for all tasks.

---

## Release & CI/CD

- **Changelog**: Automated with [git-chglog](https://github.com/git-chglog/git-chglog)
- **Release**: [goreleaser](https://goreleaser.com/) for multi-platform builds
- **CI**: GitHub Actions (`.github/workflows/go.yml`) runs lint, test, changelog, and release on tag push
- **Changelog in CI**: `task changelog-ci` runs on tag push

---

## Contributing

- Follow idiomatic Go style (see `.github/copilot-instructions.md`)
- Use Go modules (`go mod tidy`)
- Lint with `golangci-lint`
- Write and update tests for all code (`go test -cover`)
- Document code with GoDoc comments
- Keep this README up to date

---

## License

MIT License © 2023–2025 Eric N. Garcia

---

## Links

- [GitHub Repo](https://github.com/eng618/eng)
- [Releases](https://github.com/eng618/eng/releases)
- [Changelog](CHANGELOG.md)
- [Taskfile](Taskfile.yaml)
