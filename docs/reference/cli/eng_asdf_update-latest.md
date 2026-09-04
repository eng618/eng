## eng asdf update-latest

Update .tool-versions file to the latest available releases for each tool (defaults to $HOME/.tool-versions)

### Synopsis

Reads a .tool-versions file (defaults to the user level global config at $HOME/.tool-versions), queries 'asdf latest' for each plugin, and updates the file with the newest available releases.

Use the --config (-c) flag to specify a custom .tool-versions file path.
Use the --install (-i) flag to automatically install the upgraded tool versions.
Use the --yes (-y) flag to skip prompts and apply all upgrades.

```
eng asdf update-latest [flags]
```

### Options

```
  -c, --config string   Path to specific .tool-versions file (defaults to user level global config at $HOME/.tool-versions)
  -h, --help            help for update-latest
  -i, --install         Automatically run asdf install after updating .tool-versions
  -y, --yes             Skip prompts and apply all tool upgrades
```

### Options inherited from parent commands

```
  -v, --verbose   verbose output
```

### SEE ALSO

* [eng asdf](eng_asdf.md)	 - Manage asdf version manager plugins and installs

