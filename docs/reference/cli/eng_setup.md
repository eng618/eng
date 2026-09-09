## eng setup

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
eng setup [flags]
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

* [eng](eng.md)	 - A personal CLI to facilitate workflow and system maintenance.
* [eng setup asdf](eng_setup_asdf.md)	 - Setup asdf plugins from $HOME/.tool-versions
* [eng setup compaudit-fix](eng_setup_compaudit-fix.md)	 - Fix insecure directories reported by compaudit
* [eng setup dotfiles](eng_setup_dotfiles.md)	 - Setup dotfiles from your git repository
* [eng setup oh-my-zsh](eng_setup_oh-my-zsh.md)	 - Install Oh My Zsh

