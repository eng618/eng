## eng files shred

Securely delete files and directories by overwriting data

### Synopsis

Securely delete files and directories by overwriting their data multiple times
before removal. This prevents recovery of deleted data using forensic tools.

Supports multiple overwrite methods including DoD 5220.22-M (3-pass) and Gutmann (35-pass).
By default, uses a 3-pass random overwrite which is suitable for most use cases.

Examples:
  eng files shred sensitive.txt                    # Shred a single file (3 passes)
  eng files shred -r secrets/                      # Shred directory recursively
  eng files shred -p 7 -m dod file1 file2         # 7-pass DoD method
  eng files shred --dry-run secrets/              # Preview what would be shredded
  eng files shred -f -r /path/to/data             # Force (no confirmation)

```
eng files shred [paths...] [flags]
```

### Options

```
  -n, --dry-run           Preview what would be shredded without deleting
      --follow-symlinks   Follow symlinks and shred targets (default: shred symlink target)
  -f, --force             Skip confirmation prompt
  -h, --help              help for shred
  -m, --method string     Overwrite method: auto, random, zero, dod, gutmann (default "auto")
  -p, --passes int        Number of overwrite passes (0 = method default) (default 3)
  -r, --recursive         Shred directories recursively (required for directories)
      --use-system        Use system shred command on Linux (fallback to Go implementation) (default true)
      --verify            Verify overwrites by reading back (slower)
```

### Options inherited from parent commands

```
      --config string   config file (default is $HOME/.eng.yaml)
  -v, --verbose         verbose output
```

### SEE ALSO

* [eng files](eng_files.md)	 - A command for managing files

