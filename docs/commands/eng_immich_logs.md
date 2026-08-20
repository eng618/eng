## eng immich logs

View live Immich service or container logs

```
eng immich logs [flags]
```

### Options

```
  -f, --follow           Follow log stream (default true)
  -h, --help             help for logs
  -s, --service string   Filter to specific container (e.g. server, database, ml, redis)
  -n, --tail int         Number of lines to show from end of logs (default 50)
```

### Options inherited from parent commands

```
      --config string   config file (default is $HOME/.eng.yaml)
  -v, --verbose         verbose output
```

### SEE ALSO

* [eng immich](eng_immich.md)	 - Manage Immich photo stack, database, backups, and lifecycle

