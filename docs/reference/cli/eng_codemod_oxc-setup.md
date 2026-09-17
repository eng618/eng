## eng codemod oxc-setup

Setup Oxlint + Oxfmt via @gv-tech/oxc-config

### Synopsis

Install and configure Oxlint + Oxfmt for a Node.js project using @gv-tech/oxc-config.

Writes oxlint.config.ts and oxfmt.config.ts, updates package.json scripts
(lint, lint:fix, format, format:ci), configures lint-staged + Husky, and can
migrate off ESLint/Prettier.

Migration: pass --remove-eslint to uninstall the ESLint/Prettier stack and
delete legacy configs, or run without the flag to be prompted when an
ESLint/Prettier setup is detected.

```
eng codemod oxc-setup [flags]
```

### Examples

```
  eng codemod oxc-setup
  eng codemod oxc-setup --preset next
  eng codemod oxc-setup --preset vite --type-aware
  eng codemod oxc-setup --remove-eslint --yes
```

### Options

```
  -h, --help            help for oxc-setup
      --preset string   Oxc preset: auto, recommended, next, vite, react, typescript, base (default "auto")
      --remove-eslint   Remove ESLint/Prettier stack and legacy configs
      --type-aware      Enable type-aware linting (oxlint-tsgolint)
  -y, --yes             Skip confirmation prompts
```

### Options inherited from parent commands

```
      --config string   config file (default is $HOME/.eng.yaml)
      --no-color        disable colored output (also respects NO_COLOR)
  -v, --verbose         verbose output
```

### SEE ALSO

* [eng codemod](eng_codemod.md)	 - Helpers for codemods and project automation

