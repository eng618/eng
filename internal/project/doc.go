/*
Package project provides the core domain logic for managing multi-repository workspaces.

It is designed to cleanly separate the command-line interface (CLI) concerns from the
underlying business logic of fetching, pulling, syncing, and setting up projects. By
abstracting git operations behind the [RepoClient] interface, this package enables fast,
deterministic, and reliable operations across multiple repositories concurrently.

Key concepts and features:

  - **Options Structs:** Each primary operation (Setup, Sync, Pull, Fetch, List) is
    configured via an `Options` struct. These structs contain pure data, cleanly decoupling
    the package from global configuration parsers like Viper.

  - **RepoClient Abstraction:** All external side effects (e.g., executing git clone, git fetch,
    checking for dirty working trees) are abstracted behind the `RepoClient` interface.
    If no client is explicitly provided in the options, a default OS-based client is used.
    This architecture enables high test coverage without hitting the network or
    creating real repositories on disk.

  - **Concurrency & Resilience:** Operations such as Sync, Pull, and Fetch are designed
    to run in parallel using an `errgroup`, capped at a safe concurrency limit.
    Interactive prompts and network timeouts are strictly disabled to prevent
    the CLI from hanging indefinitely on network or authentication issues.

  - **Safety First:** The default workflow respects the developer's working tree.
    Pull and Sync operations skip repositories that are "dirty" (contain uncommitted changes)
    to prevent merge conflicts or interrupted rebases mid-sync. Furthermore, operations
    pull the *current* branch rather than forcing checkouts to the default branch.

Example Usage:

	opts := project.SyncOptions{
		DevPath:  "/path/to/dev",
		Projects: myConfiguredProjects,
		DryRun:   false,
	}

	// Executes a concurrent sync, pulling the latest code for all projects
	// without uncommitted changes.
	project.Sync(context.Background(), opts)
*/
package project
