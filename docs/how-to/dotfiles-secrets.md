# Back Up and Restore Dotfiles Secrets

Goal: keep machine-specific `.env` files out of git while restoring them on
any machine from Bitwarden Secrets Manager (`bws`).

## How it works

A tracked manifest (default `$HOME/bin/secrets/server.manifest`, overridable
with `--manifest`) maps managed env files to their `.example` templates and
secret keys. The Bitwarden project UUID resolves from `--project-id`,
`BWS_PROJECT_ID`, or the manifest header.

## Back up current values

```sh
eng dotfiles secrets backup
```

## Restore on a (new) machine

```sh
export BWS_ACCESS_TOKEN=...   # or export BWS_PROJECT_ID=...
eng dotfiles secrets restore
```

`eng system setup dotfiles` runs the restore automatically after install when
the manifest exists and `BWS_ACCESS_TOKEN` is set.

## Check consistency

```sh
eng dotfiles secrets doctor   # validate manifest ↔ templates ↔ secrets
eng doctor                    # overall workstation health
```

## See also

- [Set up a new machine](new-machine-setup.md)
- [Command reference: dotfiles](../reference/cli/eng_dotfiles.md)
