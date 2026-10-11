## eng proxy edit

Edit an existing proxy configuration

### Synopsis

Modify an existing proxy configuration via flags or interactively.

The target is selected by positional arg or --title (never both). The proxy
title itself is only changed with --new-title; --title never renames.

Examples:
  eng proxy edit corp --url http://proxy:8080
  eng proxy edit --title corp --new-title headquarters
  eng proxy edit 1 --no-proxy internal.corp --enable

```
eng proxy edit [name|index] [flags]
```

### Options

```
      --enable             Enable this proxy after editing
  -h, --help               help for edit
      --interactive        Use interactive prompts when missing values
      --new-title string   Rename the selected proxy
      --no-proxy string    Additional no_proxy values (comma-separated)
      --title string       Select the proxy to edit (never renames)
      --url string         Proxy address (e.g., http://host:port)
      --value string       Alias for --url
```

### Options inherited from parent commands

```
      --compact         Show compact status output (default true)
      --config string   config file (default is $HOME/.eng.yaml)
      --env             Include environment variables in status output
      --lowercase-env   Include lowercase environment vars in compact mode
      --no-color        disable colored output (also respects NO_COLOR)
  -v, --verbose         verbose output
```

### SEE ALSO

* [eng proxy](eng_proxy.md)	 - Show or configure system proxies

