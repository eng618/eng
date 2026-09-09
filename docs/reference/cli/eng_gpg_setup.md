## eng gpg setup

Setup GPG keys for signing and encryption

### Synopsis

Setup GPG keys for signing commits and encryption. This command will:
  - Prompt you for GPG key files to import
  - Import master key and subkeys
  - Set ultimate trust on the key
  - Configure Git to use your GPG key for signing
  - Optionally remove the master key (keeping only subkeys for security)
  - Configure ~/.gnupg and wrapper permissions and restart gpg-agent

```
eng gpg setup [flags]
```

### Options

```
  -h, --help   help for setup
```

### Options inherited from parent commands

```
      --config string   config file (default is $HOME/.eng.yaml)
  -v, --verbose         verbose output
```

### SEE ALSO

* [eng gpg](eng_gpg.md)	 - Manage GPG keys for commit signing and encryption

