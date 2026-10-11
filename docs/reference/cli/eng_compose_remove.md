## eng compose remove

Remove a registered compose stack

### Synopsis

Remove a user-registered compose stack from containers.stacks.

Discovered stacks under the legacy containers.path tree are not affected;
only explicitly added names can be removed.

With no arguments, an interactive picker lists the registered stacks.

Example:
  eng compose remove         # Interactive picker
  eng compose remove media

```
eng compose remove [name] [flags]
```

### Options

```
  -h, --help   help for remove
```

### Options inherited from parent commands

```
      --config string   config file (default is $HOME/.eng.yaml)
      --no-color        disable colored output (also respects NO_COLOR)
  -v, --verbose         verbose output
```

### SEE ALSO

* [eng compose](eng_compose.md)	 - Manage Docker Compose swarms and services

