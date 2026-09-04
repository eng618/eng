## eng system gpg sync

Sync updated GPG public key and expiration dates from keyservers

### Synopsis

Pull and sync updated GPG public keys and signatures from OpenPGP keyservers
and GitHub on secondary devices without needing access to your master key.

```
eng system gpg sync [flags]
```

### Options

```
      --github-user string   GitHub username to fetch public key from (e.g. eng618)
  -h, --help                 help for sync
  -k, --key-id string        GPG key ID (defaults to git user.signingkey)
      --keyserver string     Keyserver URL (default "hkps://keys.openpgp.org")
```

### Options inherited from parent commands

```
      --config string   config file (default is $HOME/.eng.yaml)
  -v, --verbose         verbose output
```

### SEE ALSO

* [eng system gpg](eng_system_gpg.md)	 - Manage GPG keys for commit signing and encryption

