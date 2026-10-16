## eng dotfiles status

check the status of your dotfiles repository

### Synopsis

This command checks the status of your local bare dotfiles repository to see if there are any local changes.

When changes exist it prints a diff --stat summary with a hint to run
'eng cfg diff' for the full patch. Use --diff to pretty-print the full
diff inline.

When run interactively without --diff and changes are detected, it
offers to show the full diff inline (same rendering as 'eng cfg diff').

```
eng dotfiles status [flags]
```

### Examples

```
  eng dotfiles status
  eng cfg status --diff
  eng cfg status --diff --staged
  eng cfg status --yes
```

### Options

```
  -d, --diff     pretty-print the diff of local changes after status
  -h, --help     help for status
      --staged   with --diff, show staged changes (git diff --cached)
  -y, --yes      automatically show the full diff when changes exist (skips the prompt)
```

### Options inherited from parent commands

```
      --config string   config file (default is $HOME/.eng.yaml)
      --no-color        disable colored output (also respects NO_COLOR)
  -v, --verbose         verbose output
```

### SEE ALSO

* [eng dotfiles](eng_dotfiles.md)	 - Manage dotfiles

