## eng fmt

Check or format code with Oxfmt using @gv-tech/oxc-config style

### Synopsis

Check or format a file or directory with Oxfmt using the shared
@gv-tech/oxc-config style, without requiring a local install.

By default (and with --check) it verifies formatting: it lists files that
would change and exits non-zero when the workspace is dirty, without
modifying anything. Pass --write to format files in place, or --diff to
also show a colored unified-style diff of the pending changes.

The formatter runs through the gv-oxfmt wrapper (bun in bun workspaces,
npx otherwise) so the shared style applies even without a local install. A
local oxfmt binary is only used as a last resort, or when forced with
--runner oxfmt. Use --config to format with a custom oxfmt config instead.

This formats JS/TS and web assets (the Oxfmt domain); Go code is formatted
with gofmt via 'task format'.

```
eng fmt [path] [flags]
```

### Examples

```
  eng fmt
  eng fmt ./web
  eng fmt src/index.ts --diff
  eng fmt --write .
  eng fmt --write src --runner bun
  eng fmt --check .
```

### Options

```
      --check           Verify formatting without modifying files (default behavior)
  -c, --config string   Custom oxfmt config file (default uses @gv-tech/oxc-config style)
      --diff            Show a colored diff of pending changes (implies check mode)
  -h, --help            help for fmt
      --runner string   JS runner: auto, npx, bun, oxfmt (default "auto")
  -w, --write           Format files in place (default is check mode)
```

### Options inherited from parent commands

```
      --no-color   disable colored output (also respects NO_COLOR)
  -v, --verbose    verbose output
```

### SEE ALSO

- [eng](eng.md) - A personal CLI to facilitate workflow and system maintenance.
