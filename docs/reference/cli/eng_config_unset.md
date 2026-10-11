## eng config unset

Remove a configuration key from the config file

### Synopsis

Removes a top-level or dotted key from the config file entirely.
Useful for dropping orphan keys flagged by validate, e.g. project_filter.

Example: eng config unset project_filter

```
eng config unset <key> [flags]
```

### Options

```
  -h, --help   help for unset
```

### Options inherited from parent commands

```
      --config string   config file (default is $HOME/.eng.yaml)
      --no-color        disable colored output (also respects NO_COLOR)
  -v, --verbose         verbose output
```

### SEE ALSO

* [eng config](eng_config.md)	 - Manage the cli's config file.

