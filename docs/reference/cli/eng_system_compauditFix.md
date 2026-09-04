## eng system compauditFix

Fix insecure directories reported by compaudit

### Synopsis

Runs 'compaudit' and applies 'chmod g-w,o-w' to any directories reported as insecure. This uses an interactive zsh so zsh functions like compaudit (from oh-my-zsh) are available.

```
eng system compauditFix [flags]
```

### Options

```
  -h, --help   help for compauditFix
```

### Options inherited from parent commands

```
      --config string   config file (default is $HOME/.eng.yaml)
  -v, --verbose         verbose output
```

### SEE ALSO

* [eng system](eng_system.md)	 - A command for managing the system

