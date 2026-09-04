## eng system proxy

Show or configure system proxies

### Synopsis

Display, switch, test, and manage multiple proxy configurations with rich visual feedback.

```
eng system proxy [flags]
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

* [eng system](eng_system.md)	 - A command for managing the system
* [eng system proxy add](eng_system_proxy_add.md)	 - Add a new proxy configuration
* [eng system proxy edit](eng_system_proxy_edit.md)	 - Edit an existing proxy configuration
* [eng system proxy export](eng_system_proxy_export.md)	 - Export proxy settings as environment variables for current shell
* [eng system proxy off](eng_system_proxy_off.md)	 - Deactivate all proxies and unset environment variables
* [eng system proxy remove](eng_system_proxy_remove.md)	 - Remove a proxy configuration
* [eng system proxy status](eng_system_proxy_status.md)	 - Show active proxy status, profiles, and shell env vars
* [eng system proxy test](eng_system_proxy_test.md)	 - Test HTTP connection through a proxy
* [eng system proxy toggle](eng_system_proxy_toggle.md)	 - Toggle proxies on or off
* [eng system proxy use](eng_system_proxy_use.md)	 - Activate a proxy configuration

