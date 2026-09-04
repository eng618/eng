## eng logs show

Show a session log (latest by default)

### Synopsis

Prints a captured session log. With no name, shows the latest run.
Names accept exact matches or unique prefixes (see 'eng logs list').

Use --tail N to show only the last N lines, or --follow to stream new
lines like 'tail -f' until interrupted.

```
eng logs show [name] [flags]
```

### Examples

```
  eng logs show
  eng logs show git-sync-all
  eng logs show --tail 50
  eng logs show system-update -f
```

### Options

```
  -f, --follow     Stream new lines until interrupted
  -h, --help       help for show
      --tail int   Show only the last N lines (0 = whole file)
```

### Options inherited from parent commands

```
      --config string   config file (default is $HOME/.eng.yaml)
  -v, --verbose         verbose output
```

### SEE ALSO

* [eng logs](eng_logs.md)	 - View session logs from verbose commands

