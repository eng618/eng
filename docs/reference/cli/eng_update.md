## eng update

Update the system and perform maintenance

### Synopsis

This command updates the system, Homebrew packages, asdf plugins, flatpak packages, and performs cleanup operations.

```
eng update [flags]
```

### Options

```
      --cleanup-timeout int   Timeout in seconds for cleanup confirmation prompt (default 60)
  -h, --help                  help for update
  -y, --yes                   Auto-approve cleanup operations without prompting
```

### Options inherited from parent commands

```
      --config string   config file (default is $HOME/.eng.yaml)
  -v, --verbose         verbose output
```

### SEE ALSO

* [eng](eng.md)	 - A personal CLI to facilitate workflow and system maintenance.
* [eng update brew](eng_update_brew.md)	 - Update Homebrew packages only
* [eng update ide](eng_update_ide.md)	 - Update or install Antigravity IDE

