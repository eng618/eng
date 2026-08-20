## eng system proxy add

Add a new proxy configuration

### Synopsis

Add a new proxy configuration with a title, proxy address, and optional bypass domains.

```
eng system proxy add [title] [url] [flags]
```

### Options

```
      --enable            Enable proxy after adding
  -h, --help              help for add
      --no-proxy string   Additional no_proxy values (comma-separated)
      --title string      Proxy configuration title
      --url string        Proxy address (e.g., http://host:port)
      --value string      Alias for --url
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

