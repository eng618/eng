## eng version

Print the version number of eng and check for updates

### Synopsis

Displays the application's version, build commit, build date, Go version,
and target OS/Architecture.

It also checks the GitHub repository (eng618/eng) for the latest official release
and compares it with the currently running version.

If a newer version is available, you can use the --update flag to attempt
an automatic upgrade via Homebrew (if installed that way) or via the
install script (if installed via curl).

```
eng version [flags]
```

### Options

```
  -h, --help     help for version
  -u, --update   Attempt to update eng to the latest version (Homebrew or install script)
```

### Options inherited from parent commands

```
      --config string   config file (default is $HOME/.eng.yaml)
  -v, --verbose         verbose output
```

### SEE ALSO

* [eng](eng.md)	 - A personal CLI to facilitate workflow and system maintenance.

