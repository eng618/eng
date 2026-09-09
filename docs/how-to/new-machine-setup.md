# Set Up a New Machine

Goal: take a fresh macOS or Linux machine to a fully provisioned dev environment
with one command (plus a few interactive prompts for keys).

## Run the full setup

```sh
eng setup
```

This runs, in order: Oh My Zsh, asdf plugins (from `~/.tool-versions`),
dotfiles install, software installation, GPG keys, and dotfiles secrets
restore (when `BWS_ACCESS_TOKEN` is set). To approve each step interactively:

```sh
eng setup --interactive
```

## Run a single step

Each step is also a standalone command for re-runs and debugging:

```sh
eng setup asdf         # install tool versions from ~/.tool-versions
eng setup dotfiles     # install dotfiles (+ secrets restore when configured)
eng setup oh-my-zsh    # install Oh My Zsh
eng ssh setup          # generate/configure GitHub SSH keys
eng gpg setup          # generate/configure GPG signing keys
```

If SSH authentication blocks a private dotfiles clone, prepare keys first:

```sh
eng ssh setup
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
- [Command reference: setup](../reference/cli/eng_setup.md)
