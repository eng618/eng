## eng kill process

Find and kill a process by PID or interactively

### Synopsis

This command finds and kills a process by its PID, or lists processes for interactive selection.

A comma-separated list of PIDs may be provided to kill multiple processes, e.g.
"eng kill process 1234,5678".

If no PID is provided or --interactive is used, it lists running processes for selection.
Requires 'ps' and 'kill' commands to be available on the system.
Primarily intended for Unix-like systems (Linux, macOS).

```
eng kill process [pid] [flags]
```

### Options

```
  -f, --filter string   Filter processes by command name
  -h, --help            help for process
  -i, --interactive     List processes interactively for selection
  -s, --signal string   Signal to send to the process (default 9 for SIGKILL) (default "9")
```

### Options inherited from parent commands

```
      --config string   config file (default is $HOME/.eng.yaml)
  -v, --verbose         verbose output
```

### SEE ALSO

* [eng kill](eng_kill.md)	 - Find and kill processes by port or PID

