## eng config secrets migrate

Move plaintext tokens from config into keychain or bitwarden

### Synopsis

Moves gitlab.token from plaintext config into the chosen store,
clears the plaintext key, and records the reference.

Example: eng config secrets migrate --to keychain --item gitlab

```
eng config secrets migrate [flags]
```

### Options

```
  -h, --help          help for migrate
      --item string   item/account name in the store (default "gitlab")
      --to string     secure store: keychain or bitwarden (default "keychain")
```

### Options inherited from parent commands

```
      --config string   config file (default is $HOME/.eng.yaml)
      --no-color        disable colored output (also respects NO_COLOR)
  -v, --verbose         verbose output
```

### SEE ALSO

* [eng config secrets](eng_config_secrets.md)	 - Manage secret storage (bitwarden, keychain, env)

