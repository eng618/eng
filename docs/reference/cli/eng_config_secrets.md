## eng config secrets

Manage secret storage (bitwarden, keychain, env)

### Synopsis

Inspect and migrate secret storage. Tokens resolve with precedence:
env vars first, then Bitwarden items, then OS keychain accounts.
Plaintext config storage is deprecated.

Example: eng config secrets migrate --to keychain --item gitlab

### Options

```
  -h, --help   help for secrets
```

### Options inherited from parent commands

```
      --config string   config file (default is $HOME/.eng.yaml)
      --no-color        disable colored output (also respects NO_COLOR)
  -v, --verbose         verbose output
```

### SEE ALSO

* [eng config](eng_config.md)	 - Manage the cli's config file.
* [eng config secrets migrate](eng_config_secrets_migrate.md)	 - Move plaintext tokens from config into keychain or bitwarden

