## eng dotfiles copy-changes

copy modified dotfiles to local git repo

### Synopsis

This command copies modified dotfiles from the worktree to the local git repository for committing.

The destination resolves as: --repo flag, explicit config
('eng config dotfiles-target-repo-path'), then dev-folder heuristics.
When nothing resolves to a git repository, it offers to locate one
interactively and persists the choice.

```
eng dotfiles copy-changes [flags]
```

### Examples

```
  eng dotfiles copy-changes
  eng dotfiles copy-changes --repo ~/Development/dotfiles
```

### Options

```
  -h, --help          help for copy-changes
      --repo string   Destination git repository path (overrides config and heuristics)
```

### Options inherited from parent commands

```
      --config string   config file (default is $HOME/.eng.yaml)
  -v, --verbose         verbose output
```

### SEE ALSO

* [eng dotfiles](eng_dotfiles.md)	 - Manage dotfiles

