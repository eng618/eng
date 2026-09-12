## eng config git-editor

Update config default editor

### Synopsis

Select the default editor used by the dashboard (e) from installed editors detected on PATH.

You may also pass the editor command directly. The dashboard falls back to $VISUAL/$EDITOR,
then agy-ide, code, and nano when no default is configured.

```
eng config git-editor [editor] [flags]
```

### Examples

```
  eng config git-editor
  eng config git-editor agy-ide
  eng config editor code
```

### Options

```
  -h, --help   help for git-editor
```

### Options inherited from parent commands

```
      --config string   config file (default is $HOME/.eng.yaml)
  -v, --verbose         verbose output
```

### SEE ALSO

* [eng config](eng_config.md)	 - Manage the cli's config file.

