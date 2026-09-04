## eng system setup dotfiles

Setup dotfiles from your git repository

### Synopsis

Setup dotfiles from your git repository. This command will:
	- Check and install prerequisites (Homebrew, Git, Bash)
	- Setup SSH keys for GitHub when required by the repository URL
  - Clone your dotfiles repository as a bare repository
  - Backup any conflicting files
  - Checkout dotfiles to your home directory
  - Initialize git submodules
	- Configure git to hide untracked files
	- Restore dotfiles secrets when manifest and BWS token are available

```
eng system setup dotfiles [flags]
```

### Options

```
  -h, --help   help for dotfiles
```

### Options inherited from parent commands

```
      --config string   config file (default is $HOME/.eng.yaml)
  -v, --verbose         verbose output
```

### SEE ALSO

* [eng system setup](eng_system_setup.md)	 - Setup development tools

