## eng config set

Set a single configuration value

### Synopsis

Sets one configuration key and validates it before writing.

Keys: email, verbose, git.dev_path, git.editor, dotfiles.repo_url,
dotfiles.branch, dotfiles.bare_repo_path, dotfiles.worktree_path,
dotfiles.target_repo_path, containers.path, antigravity.ide_download_url.

Example: eng config set email you@example.com

```
eng config set <key> <value> [flags]
```

### Options

```
  -h, --help   help for set
```

### Options inherited from parent commands

```
      --config string   config file (default is $HOME/.eng.yaml)
      --no-color        disable colored output (also respects NO_COLOR)
  -v, --verbose         verbose output
```

### SEE ALSO

* [eng config](eng_config.md)	 - Manage the cli's config file.

