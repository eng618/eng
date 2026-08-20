## eng system gpg renew

Renew or extend GPG key and subkey expiration

### Synopsis

Interactively inspect and extend expiration dates for your GPG key and subkeys.
Imports the master key from a local/cloud backup directory, updates expiration dates,
archives old key files to archive/<date-expired>/, re-exports updated secret and public
key files back to your backup folder with standard names (for cloud sync), publishes
public keys, and strips the master key locally (keeping subkeys only for security).

```
eng system gpg renew [flags]
```

### Options

```
      --duration string   Expiration duration (e.g., 1y, 2y, 6m) (default "1y")
  -h, --help              help for renew
      --keep-master       Keep master key in local keyring (do not strip to subkeys-only)
  -d, --key-dir string    Directory containing GPG key backups (default "/home/eng618/Downloads/gpg")
```

### Options inherited from parent commands

```
      --config string   config file (default is $HOME/.eng.yaml)
  -v, --verbose         verbose output
```

### SEE ALSO

* [eng system gpg](eng_system_gpg.md)	 - Manage GPG keys for commit signing and encryption

