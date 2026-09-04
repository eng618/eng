## eng compose clean

Prune unused and dangling Docker images, build cache, and volumes

### Synopsis

Prune dangling layers, unused images older than a specified duration (default: 7 days / 168h), BuildKit build cache, and optionally unused volumes to reclaim host storage.

Use the --dry-run (-n) flag to preview reclaimable space without removing data.
Use the --all (-a) flag to remove all unused images regardless of age.
Use the --volumes (-v) flag to also prune unused anonymous volumes.

```
eng compose clean [flags]
```

### Options

```
  -a, --all                 Prune all unused images regardless of age
  -b, --build-cache         Prune BuildKit / Docker build cache (default true)
  -n, --dry-run             Preview reclaimable space without deleting
  -h, --help                help for clean
  -o, --older-than string   Prune unused images older than duration (e.g. 168h, 72h) (default "168h")
      --volumes             Prune unused Docker anonymous volumes
  -y, --yes                 Skip confirmation prompt
```

### Options inherited from parent commands

```
      --config string   config file (default is $HOME/.eng.yaml)
  -v, --verbose         verbose output
```

### SEE ALSO

* [eng compose](eng_compose.md)	 - Manage Docker Compose swarms and services

