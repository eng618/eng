## eng system setup

Setup development tools

### Synopsis

Setup various development tools.
Running this command without subcommands will run all setup steps:
- Oh My Zsh
- ASDF plugins
- Dotfiles installation
- Dotfiles secrets restore (when configured)
- Software installation
- GPG keys setup (interactive)
- GPG permissions fix

```
eng system setup [flags]
```

### Options

```
  -h, --help          help for setup
  -i, --interactive   Prompt before each setup step with continue/skip/exit options
```

### Options inherited from parent commands

```
      --config string   config file (default is $HOME/.eng.yaml)
  -v, --verbose         verbose output
```

### SEE ALSO

* [eng system](eng_system.md)	 - A command for managing the system
* [eng system setup asdf](eng_system_setup_asdf.md)	 - Setup asdf plugins from $HOME/.tool-versions
* [eng system setup dotfiles](eng_system_setup_dotfiles.md)	 - Setup dotfiles from your git repository
* [eng system setup gpg](eng_system_setup_gpg.md)	 - Setup GPG keys for signing and encryption
* [eng system setup oh-my-zsh](eng_system_setup_oh-my-zsh.md)	 - Install Oh My Zsh
* [eng system setup ssh](eng_system_setup_ssh.md)	 - Setup SSH keys for GitHub access

