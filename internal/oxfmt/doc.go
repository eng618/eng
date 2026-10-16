// Package oxfmt checks and formats JavaScript/TypeScript workspaces with Oxfmt
// using the shared @gv-tech/oxc-config style, without requiring a local install.
//
// Bare invocations verify formatting (check mode); pass Options.Write to format
// files in place. Set Options.Diff to also render a colored unified-style diff
// of the pending changes.
package oxfmt
