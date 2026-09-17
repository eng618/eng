## eng config init

Guided walkthrough to configure eng step by step

### Synopsis

Walks through profile, git, dotfiles, and telemetry settings one
section at a time, showing a preview before writing anything.

Resume mid-walkthrough with --from, or preview without writing with --dry-run.

Example: eng config init --from dotfiles --dry-run

```
eng config init [flags]
```

### Options

```
      --dry-run       preview changes without writing
      --from string   resume from step: profile, git, dotfiles, telemetry
  -h, --help          help for init
```

### Options inherited from parent commands

```
      --config string   config file (default is $HOME/.eng.yaml)
      --no-color        disable colored output (also respects NO_COLOR)
  -v, --verbose         verbose output
```

### SEE ALSO

* [eng config](eng_config.md)	 - Manage the cli's config file.

