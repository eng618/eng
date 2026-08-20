## eng version

Print the version number of eng and check for updates

### Synopsis

Displays the application's version, build commit, build date, Go version,
and target OS/Architecture.

It also checks the GitHub repository (eng618/eng) for the latest official release
and compares it with the currently running version.

If a newer version is available and eng was installed via Homebrew,
you can use the --update flag to attempt an automatic upgrade.

```
eng version [flags]
```

### Options

```
  -h, --help     help for version
  -u, --update   Attempt to update eng to the latest version (requires Homebrew)
```

### Options inherited from parent commands

```
      --config string   config file (default is $HOME/.eng.yaml)
  -v, --verbose         verbose output
```

### SEE ALSO

* [eng](eng.md)	 - A personal CLI to facilitate workflow and system maintenance.

