## eng config telemetry

Manage OpenPanel telemetry and analytics settings

### Synopsis

Configure anonymous telemetry reporting to your OpenPanel instance.
Telemetry helps track command usage, execution performance, and reliability insights.

Privacy respects: The DO_NOT_TRACK=1 and ENG_TELEMETRY_DISABLED=1 environment variables
are honored globally.

```
eng config telemetry [flags]
```

### Options

```
  -h, --help   help for telemetry
```

### Options inherited from parent commands

```
      --config string   config file (default is $HOME/.eng.yaml)
  -v, --verbose         verbose output
```

### SEE ALSO

* [eng config](eng_config.md)	 - Manage the cli's config file.
* [eng config telemetry disable](eng_config_telemetry_disable.md)	 - Disable anonymous telemetry reporting
* [eng config telemetry enable](eng_config_telemetry_enable.md)	 - Enable anonymous telemetry reporting
* [eng config telemetry status](eng_config_telemetry_status.md)	 - Display active telemetry settings
* [eng config telemetry test](eng_config_telemetry_test.md)	 - Test connection to the OpenPanel telemetry instance

