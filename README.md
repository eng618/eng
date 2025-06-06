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

- **Modular CLI**: Each command is a self-contained module (dotfiles, system, codemod, ts, version, config)
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
graph TD
    A[main.go] --> B[cmd/root.go]
    B --> C1[cmd/dotfiles/]
    B --> C2[cmd/system/]
    B --> C3[cmd/codemod/]
    B --> C4[cmd/ts/]
    B --> C5[cmd/version/]
    B --> C6[cmd/config/]
    C1 --> D1[dotfiles.go]
    C1 --> D2[sync.go]
    C1 --> D3[fetch.go]
    C2 --> D4[system.go]
    C2 --> D5[kill_port.go]
    C2 --> D6[find_non_movie_folders.go]
    C2 --> D7[proxy.go]
    C2 --> D8[update.go]
    C3 --> D9[codemod.go]
    C3 --> D10[codemod_test.go]
    C4 --> D11[ts.go]
    C4 --> D12[up.go]
    C4 --> D13[down.go]
    C5 --> D14[version.go]
    C6 --> D15[config.go]
    B --> E[utils/]
    E --> F1[log/]
    E --> F2[repo/]
    E --> F3[config/]
    E --> F4[files.go]
    E --> F5[command.go]
```

---

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

- `eng dotfiles` — Manage dotfiles (sync, fetch, info)
- `eng system` — System utilities (kill-port, find-non-movie-folders, update, proxy)
- `eng codemod` — Project codemods (e.g., lint-setup)
- `eng ts` — TypeScript helpers (up, down)
- `eng version` — Show version, check for updates
- `eng config` — Show or edit config

---

## Command Reference

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
