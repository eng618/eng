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
[![Codacy Badge](https://app.codacy.com/project/badge/Grade/42051a6b1c374821b86ac29d18fb4ba2)](https://app.codacy.com/gh/ENG618/eng/dashboard?utm_source=gh&utm_medium=referral&utm_content=&utm_campaign=Badge_grade)
![GitHub Release](https://img.shields.io/github/v/release/eng618/eng)

A modern, modular CLI tool for developer automation, dotfiles management, system utilities, and project codemods. Built in Go, designed for extensibility and productivity.

## Features

- **Git Repository Management** — Bulk operations across multiple git repositories with intelligent branch detection
- **Project Management** — Organize and manage groups of related repositories as logical projects
- **Dotfiles Management** — Manage dotfiles via a bare git repo, with sync/fetch/checkout helpers
- **System Utilities** — macOS/Linux helpers (kill port, kill process, proxy, update, compaudit fix)
- **File Utilities** — Find and delete files by extension, filename, or glob pattern
- **Codemod Automation** — Project codemods (lint setup, prettier, copilot instructions)
- **Tailscale Helpers** — Bring Tailscale service up/down
- **GitLab Integration** — Manage MR rules and authentication via glab CLI
- **Config Management** — Centralized config via Viper with automatic migration
- **Auto-Update** — Version checking and Homebrew auto-update support

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
task install  # or: go install .
```

## Usage

```sh
eng [command] [flags]
```

### Quick Start

```sh
# Configure your development folder
eng config git-dev-path ~/Development

# Sync all git repositories
eng git sync-all

# Install dotfiles from your repo
eng dotfiles install

# Setup a new development machine
eng system setup

# Check version and updates
eng version
```

### Available Commands

| Command         | Description                                                        |
| --------------- | ------------------------------------------------------------------ |
| `eng git`       | Manage multiple git repositories (sync, fetch, pull, push, status) |
| `eng project`   | Manage project-based repository collections                        |
| `eng dotfiles`  | Manage dotfiles (install, sync, fetch, checkout, status)           |
| `eng system`    | System utilities (setup, kill-port, kill-process, update, proxy)   |
| `eng files`     | File utilities (find-and-delete, find-non-movie-folders)           |
| `eng codemod`   | Project codemods (lint-setup, prettier, copilot)                   |
| `eng tailscale` | Tailscale helpers (up, down) — alias: `ts`                         |
| `eng gitlab`    | GitLab integration (mr-rules, auth)                                |
| `eng version`   | Show version, check for updates                                    |
| `eng config`    | Manage CLI configuration                                           |

For detailed command documentation, see [docs/COMMANDS.md](docs/COMMANDS.md).

### Global Flags

```sh
--config string   # Config file (default: $HOME/.eng.yaml)
-v, --verbose     # Verbose output
-h, --help        # Help for any command
```

## Documentation

- [Command Reference](docs/COMMANDS.md) — Complete documentation for all commands
- [Architecture](docs/ARCHITECTURE.md) — Technical architecture and design
- [Releasing](docs/RELEASING.md) — Automated release and Homebrew publication process
- [Changelog](CHANGELOG.md) — Version history and release notes

## Support

- **Issues**: [GitHub Issues](https://github.com/eng618/eng/issues)
- **Discussions**: [GitHub Discussions](https://github.com/eng618/eng/discussions)

## Contributing

Contributions are welcome! Please follow these guidelines:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes following the [Go style guidelines](.github/copilot-instructions.md)
4. Run validation: `task validate` (runs format, lint, and tests)
5. Commit your changes using [Conventional Commits](https://www.conventionalcommits.org/)
6. Push to the branch (`git push origin feature/amazing-feature`)
7. Open a Pull Request

### Development Commands

```sh
task build      # Build the binary
task install    # Install to $GOPATH/bin
task lint       # Run golangci-lint
task test       # Run tests with coverage
task validate   # Run format + lint + test
task changelog  # Generate changelog
task release    # Create a release (goreleaser)
```

See [Taskfile.yaml](Taskfile.yaml) for all available tasks.

## Authors

- **Eric N. Garcia** — [@eng618](https://github.com/eng618)

## License

[MIT License](LICENSE) © 2023–2026 Eric N. Garcia

## Links

- [GitHub Repository](https://github.com/eng618/eng)
- [Releases](https://github.com/eng618/eng/releases)
- [Homebrew Tap](https://github.com/eng618/homebrew-eng)
