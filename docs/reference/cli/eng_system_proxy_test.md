## eng system proxy test

Test HTTP connection through a proxy

### Synopsis

Sends a test HTTP request through the specified proxy or active proxy to verify connectivity.

```
eng system proxy test [name|index] [flags]
```

### Options

```
  -h, --help            help for test
      --index int       Proxy index to test (default -1)
      --target string   Target URL to test proxy against (default "https://1.1.1.1")
      --title string    Proxy title to test
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

