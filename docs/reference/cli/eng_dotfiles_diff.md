## eng dotfiles diff

show local dotfiles changes

### Synopsis

Show unstaged and staged changes in your local bare dotfiles repository.

This is the dotfiles equivalent of running:
  git --git-dir=<repo> --work-tree=<worktree> diff

By default it shows unstaged changes. Use --staged to show staged
changes, --stat for a summary, or --name-only for just file names.
Pass optional paths to limit the diff, e.g. 'eng cfg diff -- .zshrc'.

```
eng dotfiles diff [path...] [flags]
```

### Examples

```
  eng dotfiles diff
  eng cfg diff --staged
  eng cfg diff --stat
  eng cfg diff --name-only
  eng cfg diff -- .zshrc .config/nvim
```

### Options

```
      --cached      alias for --staged (git-native spelling)
  -h, --help        help for diff
      --name-only   show only names of changed files
      --staged      show staged changes (git diff --cached)
      --stat        show diffstat summary instead of full diff
```

### Options inherited from parent commands

```
      --config string   config file (default is $HOME/.eng.yaml)
      --no-color        disable colored output (also respects NO_COLOR)
  -v, --verbose         verbose output
```

### SEE ALSO

* [eng dotfiles](eng_dotfiles.md)	 - Manage dotfiles

