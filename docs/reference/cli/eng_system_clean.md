## eng system clean

Clean and reclaim host system storage across OS and developer tools

### Synopsis

Orchestrates cross-platform system maintenance across macOS, Ubuntu, Fedora, and Debian/Raspberry Pi.

Cleans package manager caches (APT/DNF/Brew), vacuums systemd journal logs, prunes Docker container and image layers, and cleans outdated asdf tool versions.

Use the --dry-run (-n) flag to preview operations without modifying system data.
Use the --yes (-y) flag to run all available cleanup operations without prompting.

```
eng system clean [flags]
```

### Options

```
      --all-images            Prune all unused Docker images regardless of age
      --asdf                  Clean outdated asdf tool versions (default true)
      --brew                  Clean Homebrew caches (default true)
      --cleanup-timeout int   Timeout in seconds for interactive prompt (default 60)
      --docker                Include Docker container & image cleanup (default true)
  -n, --dry-run               Preview operations without modifying the system
  -h, --help                  help for clean
      --journal               Vacuum systemd journal logs (default true)
      --journal-size string   Target size for systemd journal vacuuming (default "500M")
      --older-than string     Filter age for unused Docker images (e.g. 168h) (default "168h")
      --packages              Clean OS package manager caches (APT/DNF) (default true)
  -y, --yes                   Auto-approve all cleanup operations without prompting
```

### Options inherited from parent commands

```
      --config string   config file (default is $HOME/.eng.yaml)
  -v, --verbose         verbose output
```

### SEE ALSO

* [eng system](eng_system.md)	 - A command for managing the system

