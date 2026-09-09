## eng ssh setup

Setup SSH keys for GitHub access

### Synopsis

Setup SSH keys for GitHub access. This command will:
  - Check for existing SSH keys
  - Attempt to retrieve SSH keys from Bitwarden vault
  - Generate new SSH keys if none found
  - Configure SSH config for GitHub

```
eng ssh setup [flags]
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

* [eng ssh](eng_ssh.md)	 - Manage SSH keys for GitHub access

