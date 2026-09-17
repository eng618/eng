## eng config validate

Validate the resolved configuration

### Synopsis

Loads configuration with ENG_ env overrides applied, runs central
validation, and reports one actionable error per invalid field.

Example: eng config validate

```
eng config validate [flags]
```

### Options

```
  -h, --help   help for validate
```

### Options inherited from parent commands

```
      --config string   config file (default is $HOME/.eng.yaml)
      --no-color        disable colored output (also respects NO_COLOR)
  -v, --verbose         verbose output
```

### SEE ALSO

* [eng config](eng_config.md)	 - Manage the cli's config file.

