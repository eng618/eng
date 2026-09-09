## eng kill port

Find and kill the process listening on a specific port

### Synopsis

Find the process listening on the specified port and terminate it.
Prompts for confirmation unless --yes is given. Use --dry-run to preview,
--signal to choose a gentler signal (default 9 SIGKILL), --interactive to pick from a list.

```
eng kill port [port] [flags]
```

### Examples

```
  eng kill port 3000 --dry-run
  eng kill port 3000,8080 --yes
  eng kill port --interactive
```

### Options

```
  -n, --dry-run         Preview what would be killed without killing
  -f, --filter string   Filter ports by command name
  -h, --help            help for port
  -i, --interactive     List ports interactively for selection
  -s, --signal string   Signal to send to the process (default 9 for SIGKILL) (default "9")
  -y, --yes             Skip confirmation prompt
```

### Options inherited from parent commands

```
      --config string   config file (default is $HOME/.eng.yaml)
  -v, --verbose         verbose output
```

### SEE ALSO

* [eng kill](eng_kill.md)	 - Find and kill processes by port or PID

