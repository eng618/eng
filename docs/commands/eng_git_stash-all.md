## eng git stash-all

Stash changes in all git repositories in development folder

### Synopsis

This command stashes uncommitted changes for all git repositories found in your development folder that have uncommitted changes.

```
eng git stash-all [flags]
```

### Options

```
      --dry-run          Perform a dry run without making actual changes
  -h, --help             help for stash-all
  -m, --message string   Stash message (optional)
```

### Options inherited from parent commands

```
      --config string   config file (default is $HOME/.eng.yaml)
  -c, --current         Use current working directory instead of configured development path
  -v, --verbose         verbose output
```

### SEE ALSO

* [eng git](eng_git.md)	 - Manage multiple git repositories

