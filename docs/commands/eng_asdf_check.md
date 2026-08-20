## eng asdf check

Verify that all required tool versions in .tool-versions files are installed

### Synopsis

Scans global and project .tool-versions files and verifies that every required tool version is installed on the system.

Use the --install (-i) flag to automatically install any missing tool versions.

```
eng asdf check [flags]
```

### Options

```
  -h, --help      help for check
  -i, --install   Automatically install missing tool versions
      --no-scan   Disable scanning development directories
```

### Options inherited from parent commands

```
      --config string   config file (default is $HOME/.eng.yaml)
  -v, --verbose         verbose output
```

### SEE ALSO

* [eng asdf](eng_asdf.md)	 - Manage asdf version manager plugins and installs

