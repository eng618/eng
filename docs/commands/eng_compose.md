## eng compose

Manage Docker Compose swarms and services

### Synopsis

Audit, inspect, start, stop, pull, and monitor Docker Compose service stacks.

```
eng compose [flags]
```

### Examples

```
  eng compose list
  eng compose status --all
  eng compose up media -e dev
```

### Options

```
  -h, --help   help for compose
```

### Options inherited from parent commands

```
      --config string   config file (default is $HOME/.eng.yaml)
  -v, --verbose         verbose output
```

### SEE ALSO

* [eng](eng.md)	 - A personal CLI to facilitate workflow and system maintenance.
* [eng compose clean](eng_compose_clean.md)	 - Prune unused and dangling Docker images, build cache, and volumes
* [eng compose down](eng_compose_down.md)	 - Spin down one or more Compose stacks
* [eng compose list](eng_compose_list.md)	 - List discovered Docker Compose stacks
* [eng compose logs](eng_compose_logs.md)	 - View logs from a Compose stack
* [eng compose pull](eng_compose_pull.md)	 - Pull latest service images for Compose stacks
* [eng compose status](eng_compose_status.md)	 - Show status of Compose stacks and services
* [eng compose up](eng_compose_up.md)	 - Spin up one or more Compose stacks

