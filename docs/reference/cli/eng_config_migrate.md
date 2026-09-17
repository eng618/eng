## eng config migrate

Migrate the config file to the current schema version

### Synopsis

Migrates legacy keys to the current schema, deletes renamed keys,
stamps the version, and writes a timestamped backup before changing anything.

Use --check to preview without writing.

Example: eng config migrate --check

```
eng config migrate [flags]
```

### Options

```
      --check   preview migration without writing
  -h, --help    help for migrate
```

### Options inherited from parent commands

```
      --config string   config file (default is $HOME/.eng.yaml)
      --no-color        disable colored output (also respects NO_COLOR)
  -v, --verbose         verbose output
```

### SEE ALSO

* [eng config](eng_config.md)	 - Manage the cli's config file.

