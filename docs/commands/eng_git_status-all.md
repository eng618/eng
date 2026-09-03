## eng git status-all

Check status of all git repositories in development folder

### Synopsis

Check working-tree status across all git repositories in your development folder. Use --current to scan the current directory.

```
eng git status-all [flags]
```

### Examples

```
  eng git status-all
  eng git status-all --current
```

### Options

```
  -h, --help   help for status-all
```

### Options inherited from parent commands

```
      --config string   config file (default is $HOME/.eng.yaml)
  -c, --current         Use current working directory instead of configured development path
  -v, --verbose         verbose output
```

### SEE ALSO

* [eng git](eng_git.md)	 - Manage multiple git repositories

