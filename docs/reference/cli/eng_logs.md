## eng logs

View session logs from verbose commands

### Synopsis

Verbose commands (git bulk syncs, system updates) write full detail logs
to a file while the terminal shows only a clean summary.

Use these subcommands to inspect past runs without re-running them.

```
eng logs [flags]
```

### Examples

```
  eng logs              # List recent session logs
  eng logs show         # Show the latest session log
  eng logs show git-sync-all --tail 50
  eng logs clean        # Delete all session logs
```

### Options

```
  -h, --help   help for logs
```

### Options inherited from parent commands

```
      --config string   config file (default is $HOME/.eng.yaml)
  -v, --verbose         verbose output
```

### SEE ALSO

* [eng](eng.md)	 - A personal CLI to facilitate workflow and system maintenance.
* [eng logs clean](eng_logs_clean.md)	 - Delete session logs
* [eng logs list](eng_logs_list.md)	 - List recent session logs
* [eng logs show](eng_logs_show.md)	 - Show a session log (latest by default)

