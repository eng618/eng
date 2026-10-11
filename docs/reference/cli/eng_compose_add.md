## eng compose add

Register a named compose stack at a local root path

### Synopsis

Register a named compose stack pointing at an arbitrary local directory.

The compose file is autodetected inside the directory (docker-compose.yml/yaml,
compose.yml/yaml). The registration is stored under containers.stacks in the
config file and is included in list, status, up, down, pull, and logs.

With no arguments, an interactive wizard prompts for the name and path.

Example:
eng compose add # Interactive wizard
eng compose add media ~/Development/homelab/media

```
eng compose add [name] [path] [flags]
```

### Options

```
  -h, --help   help for add
```

### Options inherited from parent commands

```
      --config string   config file (default is $HOME/.eng.yaml)
      --no-color        disable colored output (also respects NO_COLOR)
  -v, --verbose         verbose output
```

### SEE ALSO

- [eng compose](eng_compose.md) - Manage Docker Compose swarms and services
