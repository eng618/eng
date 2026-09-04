## eng config dotfiles-target-repo-path

Show or set the dotfiles target repo path for copy-changes

### Synopsis

Show or set the explicit destination git repository used by 'eng dotfiles copy-changes'.

With a path argument, validates that it exists and saves it, so future copies
no longer depend on dev-folder heuristics. Without arguments, prints the
currently effective path and where it came from.

```
eng config dotfiles-target-repo-path [path] [flags]
```

### Examples

```
  eng config dotfiles-target-repo-path
  eng config dotfiles-target-repo-path ~/Development/dotfiles
```

### Options

```
  -h, --help   help for dotfiles-target-repo-path
```

### Options inherited from parent commands

```
      --config string   config file (default is $HOME/.eng.yaml)
  -v, --verbose         verbose output
```

### SEE ALSO

* [eng config](eng_config.md)	 - Manage the cli's config file.

