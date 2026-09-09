## eng proxy

Show or configure system proxies

### Synopsis

Display, switch, test, and manage multiple proxy configurations with rich visual feedback.

```
eng proxy [flags]
```

### Options

```
      --compact         Show compact status output (default true)
      --env             Include environment variables in status output
  -h, --help            help for proxy
      --lowercase-env   Include lowercase environment vars in compact mode
```

### Options inherited from parent commands

```
      --config string   config file (default is $HOME/.eng.yaml)
  -v, --verbose         verbose output
```

### SEE ALSO

* [eng](eng.md)	 - A personal CLI to facilitate workflow and system maintenance.
* [eng proxy add](eng_proxy_add.md)	 - Add a new proxy configuration
* [eng proxy edit](eng_proxy_edit.md)	 - Edit an existing proxy configuration
* [eng proxy export](eng_proxy_export.md)	 - Export proxy settings as environment variables for current shell
* [eng proxy off](eng_proxy_off.md)	 - Deactivate all proxies and unset environment variables
* [eng proxy remove](eng_proxy_remove.md)	 - Remove a proxy configuration
* [eng proxy status](eng_proxy_status.md)	 - Show active proxy status, profiles, and shell env vars
* [eng proxy test](eng_proxy_test.md)	 - Test HTTP connection through a proxy
* [eng proxy toggle](eng_proxy_toggle.md)	 - Toggle proxies on or off
* [eng proxy use](eng_proxy_use.md)	 - Activate a proxy configuration

