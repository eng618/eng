# Releasing and Homebrew Publication

This document describes the automated process for releasing new versions of `eng` and updating the Homebrew formula.

## Overview

The release process is semi-automated using **GitHub Actions** and **Google's Release Please**. 

1.  **Release Please** tracks commits on the `main` branch and maintains a "Release PR".
2.  When the Release PR is merged, it automatically creates a GitHub Release and a Git tag.
3.  The creation of a tag triggers the **Homebrew Publication** workflow.

## Release Components

### 1. Release Please Workflow (`release-please.yml`)

- **Trigger**: Pushes to the `main` branch.
- **Actions**:
  - Maintains a Release PR (e.g., `chore: release 0.17.5`) with an updated `CHANGELOG.md` and version.
  - Upon merging the PR, it creates a new GitHub Release and tag.

### 2. Homebrew Publication Workflow (`publish-to-homebrew.yml`)

This workflow is triggered when a new tag `v*` is created.

- **Job: `build` (GoReleaser)**:
  - Builds binaries for multiple platforms.
  - Updates the GitHub Release with the compiled artifacts.
- **Job: `publish` (Homebrew Update)**:
  - Extracts checksums from GoReleaser artifacts.
  - Updates the [eng618/homebrew-eng](https://github.com/eng618/homebrew-eng) tap repository.

## Requirements

### Conventional Commits
To allow Release Please to determine the next version number and generate the changelog, you **must** use [Conventional Commits](https://www.conventionalcommits.org/):
- `feat: ...` for new features (minor version bump).
- `fix: ...` for bug fixes (patch version bump).
- `feat!: ...` or `BREAKING CHANGE: ...` for breaking changes (major version bump).

### Secrets
- `RELEASE_PLEASE_TOKEN`: A Personal Access Token (PAT) with `repo` and `workflow` scopes. This is used for:
  1.  **Release Please**: To maintain PRs and push tags that trigger subsequent workflows.
  2.  **Homebrew Publication**: To push changes to the `eng618/homebrew-eng` tap repository.
- `CODACY_PROJECT_TOKEN`: (Optional but recommended) Used by the Go CI workflow to report code coverage to Codacy.

## How to Release a New Version

1.  **Merge changes to `main`**: Ensure your commits follow the Conventional Commits format.
2.  **Wait for Release PR**: A PR titled `chore: release X.Y.Z` will be automatically opened or updated by a bot.
3.  **Approve and Merge**: Once you are ready to release, merge this PR.
4.  **Automation handles the rest**: Merging will create the tag, which triggers the build and publication to Homebrew.

## Troubleshooting

### Release PR not Updating
Ensure your commits on `main` follow the Conventional Commits prefix. If no "feature" or "fix" commits are detected since the last release, a new PR might not be created.

### Homebrew Workflow not Triggered
If the tag is created but the `Publish to Homebrew` workflow doesn't start, verify that `release-please.yml` is using the `RELEASE_PLEASE_TOKEN` (PAT) instead of the default `GITHUB_TOKEN`.
