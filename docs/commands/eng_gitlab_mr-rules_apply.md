## eng gitlab mr-rules apply

Apply merge request rules from a JSON file

```
eng gitlab mr-rules apply [flags]
```

### Options

```
      --dry-run             Print the glab command without executing
  -h, --help                help for apply
      --hose string         Alias for --host
      --host string         GitLab host (e.g., gitlab.com)
      --project string      GitLab project path (e.g., group/subgroup/repo)
      --rules string        Path to MR rules JSON file
      --token-item string   Bitwarden item name containing a GitLab token
```

### Options inherited from parent commands

```
      --config string   config file (default is $HOME/.eng.yaml)
  -v, --verbose         verbose output
```

### SEE ALSO

* [eng gitlab mr-rules](eng_gitlab_mr-rules.md)	 - Manage GitLab merge request rules

