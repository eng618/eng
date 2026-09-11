## eng sysinfo

Show system diagnostics (CPU, memory, disk, network)

### Synopsis

Collects best-effort system diagnostics and prints them as a formatted table.

Works across Raspberry Pi, Fedora, Ubuntu, and macOS. Fields that cannot be
determined render as "n/a", so the command always succeeds.

```
eng sysinfo [flags]
```

### Examples

```
  eng sysinfo
  eng sysinfo --output json
  eng sysinfo --output yaml --timeout 5s
```

### Options

```
  -h, --help               help for sysinfo
  -o, --output string      Output format: table, json, or yaml (default "table")
      --timeout duration   Overall timeout for diagnostics collection (default 8s)
```

### Options inherited from parent commands

```
      --config string   config file (default is $HOME/.eng.yaml)
  -v, --verbose         verbose output
```

### SEE ALSO

* [eng](eng.md)	 - A personal CLI to facilitate workflow and system maintenance.

