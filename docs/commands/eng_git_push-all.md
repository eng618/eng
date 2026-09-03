## eng git push-all

Push all git repositories in development folder

### Synopsis

Push commits for all git repositories in your development folder that have unpushed commits. Use --dry-run to preview, --force only with --yes confirmation for force-with-lease pushes.

```
eng git push-all [flags]
```

### Examples

```
  eng git push-all --dry-run
  eng git push-all
  eng git push-all --force --yes
```

### Options

```
  -n, --dry-run   Preview what would be pushed without making changes
      --force     Force push with --force-with-lease (requires confirmation unless --yes)
  -h, --help      help for push-all
  -y, --yes       Skip confirmation prompts
```

### Options inherited from parent commands

```
      --config string   config file (default is $HOME/.eng.yaml)
  -c, --current         Use current working directory instead of configured development path
  -v, --verbose         verbose output
```

### SEE ALSO

* [eng git](eng_git.md)	 - Manage multiple git repositories

