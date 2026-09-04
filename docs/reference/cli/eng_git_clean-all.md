## eng git clean-all

Clean untracked files in all git repositories in development folder

### Synopsis

Preview untracked files across all repos, confirm, then clean. Use --dry-run to preview only, --force with --yes to skip confirmation.

```
eng git clean-all [flags]
```

### Examples

```
  eng git clean-all --dry-run
  eng git clean-all
  eng git clean-all --force --yes -d
```

### Options

```
  -d, --directories   Also remove untracked directories
  -n, --dry-run       Preview what would be cleaned without making changes
      --force         Skip confirmation prompt (use with --yes in scripts)
  -h, --help          help for clean-all
  -y, --yes           Skip confirmation prompt
```

### Options inherited from parent commands

```
      --config string   config file (default is $HOME/.eng.yaml)
  -c, --current         Use current working directory instead of configured development path
  -v, --verbose         verbose output
```

### SEE ALSO

* [eng git](eng_git.md)	 - Manage multiple git repositories

