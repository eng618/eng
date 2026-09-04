# Set Up a New Machine

Goal: take a fresh macOS or Linux machine to a fully provisioned dev environment
with one command (plus a few interactive prompts for keys).

## Run the full setup

```sh
eng system setup
```

This runs, in order: Oh My Zsh, asdf plugins (from `~/.tool-versions`),
dotfiles install, software installation, GPG keys, and dotfiles secrets
restore (when `BWS_ACCESS_TOKEN` is set). To approve each step interactively:

```sh
eng system setup --interactive
```

## Run a single step

Each step is also a standalone command for re-runs and debugging:

```sh
eng system setup asdf         # install tool versions from ~/.tool-versions
eng system setup dotfiles     # install dotfiles (+ secrets restore when configured)
eng system setup oh-my-zsh    # install Oh My Zsh
eng system setup ssh          # generate/configure GitHub SSH keys
eng system setup gpg          # generate/configure GPG signing keys
```

If SSH authentication blocks a private dotfiles clone, prepare keys first:

```sh
eng system setup ssh
eng dotfiles install
```

## Verify

```sh
eng doctor                    # tools, paths, dotfiles, telemetry, version
eng project setup             # clone any missing project repos
```

## See also

- [Getting started](../tutorials/getting-started.md)
- [Back up and restore dotfiles secrets](dotfiles-secrets.md)
- [Command reference: system](../reference/cli/eng_system.md)
