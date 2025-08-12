# Copilot Custom Instructions

<!-- Use this file to provide workspace-specific custom instructions to Copilot. For more details, visit https://code.visualstudio.com/docs/copilot/copilot-customization#_use-a-githubcopilotinstructionsmd-file -->

## General

- Keep the README.md file up to date with the latest information about the project.
- Use clear and concise language in comments and documentation.

## Go

- Always write idiomatic Go code.
- Use the Go standard library whenever possible.
- Use the `go` command to run tests, build, and install packages.
- Use `golangci-lint` to format code.
- Use `golangci-lint run` to check for linting issues.
- Use `go test` to run tests.
- Use `go mod` to manage dependencies.

- Always write tests for new code.
- Always update tests when modifying existing code.
- Use `go test -cover` to check test coverage.
- Use `go vet` to check for common mistakes.
- Use `go doc` to generate documentation.

## git Commit messages

- Use conventional commit messages
- Follow the commit message guidelines
- See https://www.conventionalcommits.org/en/v1.0.0/
- Start commit messages with a type (e.g., feat, fix, docs, style, refactor, test, chore)
- Optionally include a scope in parentheses after the type (e.g., feat(parser): ...)
- Use a short, imperative description after the type/scope
- Separate the body from the header with a blank line if additional context is needed
- Reference issues or pull requests in the footer if applicable
- Mention breaking changes in the footer if present
