## eng system proxy remove

Remove a proxy configuration

### Synopsis

Deletes a stored proxy configuration profile.

```
eng system proxy remove [name|index] [flags]
```

### Options

```
  -h, --help           help for remove
      --index int      Proxy index to remove (default -1)
      --title string   Proxy title to remove
```

### Options inherited from parent commands

```
      --compact         Show compact status output (default true)
      --config string   config file (default is $HOME/.eng.yaml)
      --env             Include environment variables in status output
      --lowercase-env   Include lowercase environment vars in compact mode
  -v, --verbose         verbose output
```

### SEE ALSO

* [eng system proxy](eng_system_proxy.md)	 - Show or configure system proxies

