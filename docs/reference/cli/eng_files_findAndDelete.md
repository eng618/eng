## eng files findAndDelete

Find and delete files of selected types, or list extensions

### Synopsis

Recursively scan the provided directory for files of types selected by the user
and delete them after an interactive confirmation. Use --list-extensions to list
all file extensions in the directory instead. Use --filename to target a specific
filename, --glob for glob patterns, or --ext for file extensions.


```
eng files findAndDelete [directory] [flags]
```

### Options

```
  -e, --ext string        File extension to match (e.g., '.json'). Bypasses extension selection.
  -f, --filename string   Specific filename to match (e.g., 'package.json'). Bypasses extension selection.
  -g, --glob string       Glob pattern to match files (e.g., '*.bak'). Bypasses extension selection.
  -h, --help              help for findAndDelete
  -l, --list-extensions   List all file extensions in the directory
```

### Options inherited from parent commands

```
      --config string   config file (default is $HOME/.eng.yaml)
  -v, --verbose         verbose output
```

### SEE ALSO

* [eng files](eng_files.md)	 - A command for managing files

