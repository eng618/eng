# Releasing and Homebrew Publication

This document describes the automated process for releasing new versions of `eng` and updating the Homebrew formula.

## Overview

The release process is fully automated using GitHub Actions. It is triggered whenever a tag matching `v*` or `[0-9]*` is pushed to the repository.

The workflow is defined in [`.github/workflows/publish-to-homebrew.yml`](../.github/workflows/publish-to-homebrew.yml).

## Release Workflow

The workflow consists of two main jobs: `build` and `publish`.

### 1. The Build Job (`build`)

- **Tool**: Uses [GoReleaser](https://goreleaser.com/).
- **Actions**:
  - Build binaries for Darwin (macOS) and Linux across multiple architectures (`amd64`, `arm64`).
  - Creates a GitHub Release with the tag name.
  - Generates checksums for all artifacts.
  - Uploads the resulting `dist/` directory as a GitHub Action artifact for the next job.

### 2. The Publish Job (`publish`)

This job specifically handles updating the Homebrew tap at [eng618/homebrew-eng](https://github.com/eng618/homebrew-eng).

- **Checkout**: It checks out the main repository to access the update tools.
- **Artifacts**: It downloads the `dist/` artifact created in the `build` job.
- **Checksum Extraction**: It parses the `artifacts.json` produced by GoReleaser to extract the SHA256 checksums for each platform.
- **Go Script**: It runs [`tools/homebrew-update.go`](../tools/homebrew-update.go) which performs the following:
  - Clones the Homebrew tap repository.
  - Generates a new Ruby formula (`eng.rb`) using a template and the extracted checksums/version.
  - Commits and pushes the change back to the tap repository using a Personal Access Token (`PAT_TOKEN`).

## Requirements

### Personal Access Token (`PAT_TOKEN`)

The `publish` job requires a secret named `PAT_TOKEN`. This token must have `repo` scope permissions for the target Homebrew tap repository (`eng618/homebrew-eng`), as the default `GITHUB_TOKEN` only has permissions for the current repository.

## How to Trigger a Release

To release a new version:

1. Ensure all changes are committed and pushed to `main`.
2. Create a new tag:

   ```bash
   git tag v1.3.4
   ```

3. Push the tag:

   ```bash
   git push origin v1.3.4
   ```

The GitHub Actions will take care of the rest. You can monitor the progress in the **Actions** tab of the repository.

## Troubleshooting

### "no such file or directory" for Go script

If the `publish` job fails to find `../tools/homebrew-update.go`, ensure that the `Checkout repository` step is present in the `publish` job. Since jobs run on fresh runners, the repository is not automatically available unless explicitly checked out.
