# Command Reference

Complete documentation for all `eng` CLI commands and their options.

## Table of Contents

- [Git Repository Management](#git-repository-management)
- [Project Management](#project-management)
- [Dotfiles Management](#dotfiles-management)
- [System Utilities](#system-utilities)
- [File Utilities](#file-utilities)
- [Codemod Tools](#codemod-tools)
- [Tailscale](#tailscale)
- [GitLab Integration](#gitlab-integration)
- [Version](#version)
- [Config](#config)

---

## Git Repository Management

Manage multiple git repositories in your development folder with a comprehensive set of commands. All commands support the `--current` flag to operate on the current directory instead of the configured development path.

### Setup

```sh
# Configure your development folder path
eng config git-dev-path /path/to/your/dev/folder
```

### Repository Operations

| Command                                     | Description                                           |
| ------------------------------------------- | ----------------------------------------------------- |
| `eng git sync-all [--current] [--dry-run]`  | Fetch and pull with rebase across all repositories    |
| `eng git fetch-all [--current] [--dry-run]` | Fetch latest changes from remote for all repositories |
| `eng git pull-all [--current] [--dry-run]`  | Pull latest changes with rebase for all repositories  |
| `eng git push-all [--current] [--dry-run]`  | Push local changes to remote for all repositories     |
| `eng git status-all [--current]`            | Show git status for all repositories                  |
| `eng git list [--current]`                  | List all git repositories found                       |
| `eng git branch-all [--current]`            | Show current branch for all repositories              |
| `eng git stash-all [--current] [--dry-run]` | Stash changes in all repositories                     |
| `eng git clean-all [--current] [--dry-run]` | Clean untracked files in all repositories             |

### Flags

- `--current` — Use current working directory instead of configured development path
- `--dry-run` — Show what would be done without making changes (where applicable)

---

## Project Management

Manage project-based repository collections. A project is a logical grouping of related repositories (e.g., microservices, shared infrastructure).

### Setup

```sh
# Configure your development folder path (required)
eng config git-dev-path /path/to/your/dev/folder
```

### Project Structure

Projects are stored in your development folder with each project having its own subdirectory:

```
~/Development/
  MyProject/
    api/
    web/
    shared/
  Infrastructure/
    core/
    auth/
```

### Commands

| Command              | Description                                              |
| -------------------- | -------------------------------------------------------- |
| `eng project --info` | Show current project configuration                       |
| `eng project list`   | List configured projects and their repositories          |
| `eng project add`    | Add a new project or repository to configuration         |
| `eng project remove` | Remove a project or repository from configuration        |
| `eng project setup`  | Setup project directories and clone missing repositories |
| `eng project fetch`  | Fetch updates for all project repositories               |
| `eng project pull`   | Pull updates for all project repositories                |
| `eng project sync`   | Sync all project repositories (fetch + pull)             |

### Flags

- `--project <name>` / `-p` — Filter operations to a specific project
- `--dry-run` — Show what would be done without making changes

---

## Dotfiles Management

Manage your dotfiles with a bare git repository approach. This allows you to version control your home directory configuration files without the overhead of a traditional git repository.

### Install

```sh
# First-time setup: Install dotfiles from your repository
eng dotfiles install
```

This command will:

- Check and install prerequisites (Homebrew, Git, Bash)
- Verify you have a valid SSH key at `~/.ssh/github`
- Prompt for and save your dotfiles repository URL and branch
- Clone your repository as a bare repository using go-git (default: `~/.eng-cfg`)
- Backup any conflicting files to a timestamped directory
- Checkout dotfiles to your home directory
- Initialize git submodules
- Configure git to hide untracked files

### System Setup Integration

```sh
# Set up dotfiles as part of new system setup
eng system setup dotfiles
```

This checks prerequisites and then guides you through the installation process.

### Configuration

```sh
# Configure dotfiles settings
eng config dotfiles-repo-url          # Set repository URL
eng config dotfiles-branch            # Set branch (main/work/server)
eng config dotfiles-bare-repo-path    # Set bare repo location
eng config dotfiles-repo              # Set dotfiles repo path
```

### Commands

| Command                     | Description                                  |
| --------------------------- | -------------------------------------------- |
| `eng dotfiles --info`       | Show current dotfiles config                 |
| `eng dotfiles install`      | First-time dotfiles installation             |
| `eng dotfiles sync`         | Fetch and pull latest dotfiles               |
| `eng dotfiles fetch`        | Fetch latest dotfiles without merging        |
| `eng dotfiles checkout`     | Checkout files from the bare repository      |
| `eng dotfiles copy-changes` | Copy modified dotfiles to local git repo     |
| `eng dotfiles status`       | Check the status of your dotfiles repository |

### Using the `cfg` Alias

After installation, you can also use the `cfg` alias (if configured in your dotfiles):

```sh
cfg status              # Check dotfiles status
cfg add ~/.vimrc        # Stage a file
cfg commit -m "Update"  # Commit changes
cfg push                # Push changes
cfg pull                # Pull changes
```

---

## System Utilities

System utilities for macOS and Linux, including developer setup automation.

### Setup Commands

| Command                      | Description                                             |
| ---------------------------- | ------------------------------------------------------- |
| `eng system setup`           | Run all setup steps (Oh My Zsh, ASDF, dotfiles)         |
| `eng system setup asdf`      | Setup asdf plugins from `$HOME/.tool-versions`          |
| `eng system setup dotfiles`  | Setup dotfiles (checks prerequisites then runs install) |
| `eng system setup oh-my-zsh` | Install Oh My Zsh                                       |
| `eng system setup ssh`       | Setup SSH keys for GitHub access                        |

### System Utilities

| Command                        | Description                                    |
| ------------------------------ | ---------------------------------------------- |
| `eng system killPort <port>`   | Kill process on a port                         |
| `eng system killProcess [pid]` | Kill a process by PID or interactively         |
| `eng system compauditFix`      | Fix insecure directories reported by compaudit |
| `eng system update`            | Update system packages                         |
| `eng system proxy`             | Manage proxy settings                          |

### killProcess Flags

- `--interactive` / `-i` — List processes for selection
- `--filter <name>` / `-f` — Filter processes by command name
- `--signal <signal>` / `-s` — Signal to send (default: 9/SIGKILL)

---

## File Utilities

File management utilities for finding and cleaning up files.

### Commands

| Command                                     | Description                   |
| ------------------------------------------- | ----------------------------- |
| `eng files findAndDelete [directory]`       | Find and delete files by type |
| `eng files findNonMovieFolders [--dry-run]` | Find/delete non-movie folders |

### findAndDelete Flags

- `--list-extensions` / `-l` — List all file extensions in directory
- `--filename <name>` / `-f` — Match specific filename (e.g., `package.json`)
- `--ext <extension>` / `-e` — Match file extension (e.g., `.json`)
- `--glob <pattern>` / `-g` — Match glob pattern (e.g., `*.bak`)

---

## Codemod Tools

Project automation and setup helpers for various development environments.

### Commands

| Command                         | Description                                                                |
| ------------------------------- | -------------------------------------------------------------------------- |
| `eng codemod lint-setup`        | Setup lint/format (eslint, prettier, husky, lint-staged) in JS/TS projects |
| `eng codemod prettier [path]`   | Format code with prettier using @eng618/prettier-config                    |
| `eng codemod copilot [--force]` | Create base custom Copilot instructions file                               |

### copilot Command Details

Creates a `.github/copilot-instructions.md` file with:

- Validates you're in a Git repository (use `--force` to bypass)
- Creates `.github/` directory if it doesn't exist
- Won't overwrite existing files
- Includes comprehensive template with code quality guidelines

---

## Tailscale

Helpers for managing the Tailscale VPN service.

| Command              | Description                                                  |
| -------------------- | ------------------------------------------------------------ |
| `eng tailscale up`   | Bring up the Tailscale service (runs `sudo tailscale up`)    |
| `eng tailscale down` | Take down the Tailscale service (runs `sudo tailscale down`) |

**Alias:** `eng ts up`, `eng ts down`

---

## GitLab Integration

Commands that integrate with GitLab using the `glab` CLI.

### MR Rules

Apply Merge Request rules to a project using a JSON rules file:

```sh
# Example rules.json (see also docs/gitlab-rules.example.json)
cat > rules.json <<'JSON'
{
  "schemaVersion": "1",
  "mergeMethod": "ff",
  "deleteSourceBranch": true,
  "requireSquash": true,
  "pipelinesMustSucceed": true,
  "allowSkippedAsSuccess": true,
  "allThreadsMustResolve": true
}
JSON

# Apply to current repo's GitLab project (auto-detect host/project)
eng gitlab mr-rules apply --rules rules.json

# Explicit host/project
eng gitlab mr-rules apply --rules rules.json --host gitlab.com --project group/sub/repo

# Dry run
eng gitlab mr-rules apply --rules rules.json --dry-run
```

Generate rules interactively:

```sh
# Prompt for each option and write to a file
eng gitlab mr-rules init --output gitlab-rules.json

# Use defaults without prompting
eng gitlab mr-rules init --yes --output gitlab-rules.json
```

### Authentication

Authentication options:

- Set `GITLAB_TOKEN` in your environment; or
- Store a token in Bitwarden and reference it via config `gitlab.tokenItem` or `--token-item <itemName>`.
  The CLI will unlock Bitwarden (prompting if needed), read the token from the item's password or a field named `token`, and export it for the `glab` call.

Config keys (optional):

- `gitlab.host` — default host (e.g., `gitlab.com`)
- `gitlab.project` — default project path (e.g., `group/sub/repo`)
- `gitlab.tokenItem` — Bitwarden item name holding token

### Auth Commands

```sh
# Save a token into Bitwarden and set defaults
printf "%s" "$GITLAB_TOKEN" | eng gitlab auth set \
  --stdin \
  --token-item "eng/gitlab-token" \
  --host gitlab.com \
  --project group/sub/repo

# Or later, just set the defaults without re-saving a token
eng gitlab auth set --host gitlab.com --project group/sub/repo --token-item "eng/gitlab-token"

# Show effective defaults and token source (no secrets)
eng gitlab auth show

# Validate token and project access (no writes)
eng gitlab auth doctor

# Force-check a specific host/project and run quietly
eng gitlab auth doctor --host gitlab.example.com --project group/sub/repo --quiet
```

---

## Version

| Command                | Description                                      |
| ---------------------- | ------------------------------------------------ |
| `eng version`          | Show version, build info, check for updates      |
| `eng version --update` | Auto-update via Homebrew (if installed that way) |

---

## Config

Manage CLI configuration stored at `$HOME/.eng.yaml`.

| Command                              | Description                                  |
| ------------------------------------ | -------------------------------------------- |
| `eng config`                         | Show current config                          |
| `eng config edit`                    | Edit config file                             |
| `eng config git-dev-path`            | Set development folder path for git commands |
| `eng config dotfiles-repo-url`       | Set dotfiles repository URL                  |
| `eng config dotfiles-branch`         | Set dotfiles branch                          |
| `eng config dotfiles-bare-repo-path` | Set bare repo location                       |
| `eng config dotfiles-repo`           | Set dotfiles repo path                       |
| `eng config email`                   | Set email address                            |
| `eng config verbose`                 | Set verbose output default                   |

---

## Global Flags

These flags are available for all commands:

- `--config string` — Config file (default is `$HOME/.eng.yaml`)
- `-v, --verbose` — Verbose output
- `-h, --help` — Help for any command
