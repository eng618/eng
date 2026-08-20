## eng gitlab auth set

Configure GitLab token and defaults

### Synopsis

Save or update a GitLab token in Bitwarden and set defaults (host, project, token item) in eng config.

```
eng gitlab auth set [flags]
```

### Options

```
  -h, --help                help for set
      --host string         Default GitLab host to save in config (e.g., gitlab.com)
      --notes string        Optional notes to save with the Bitwarden item
      --project string      Default GitLab project path to save in config (e.g., group/sub/repo)
      --stdin               Read token from STDIN
      --token string        GitLab token to save into Bitwarden (discouraged; prefer stdin)
      --token-item string   Bitwarden item name to store/find the GitLab token
```

### Options inherited from parent commands

```
      --config string   config file (default is $HOME/.eng.yaml)
  -v, --verbose         verbose output
```

### SEE ALSO

* [eng gitlab auth](eng_gitlab_auth.md)	 - Manage GitLab authentication for eng

