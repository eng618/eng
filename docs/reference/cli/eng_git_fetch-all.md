## eng git fetch-all

Fetch all git repositories in development folder

### Synopsis

This command fetches updates from remote for all git repositories found in your development folder.

Use --force to overwrite local tags when remotes move them
(git fetch --all --prune --force).

```
eng git fetch-all [flags]
```

### Options

```
      --dry-run   Perform a dry run without making actual changes
      --force     Force overwrite local tags on fetch conflicts (git fetch --force)
  -h, --help      help for fetch-all
```

### Options inherited from parent commands

```
      --config string   config file (default is $HOME/.eng.yaml)
  -c, --current         Use current working directory instead of configured development path
  -v, --verbose         verbose output
```

### SEE ALSO

* [eng git](eng_git.md)	 - Manage multiple git repositories

