# Testing in the `eng` CLI

This document outlines the testing strategy, architecture, and standards used throughout the `eng` CLI. We follow the **BET (Benchmarks, Examples, Tests)** testing methodology to verify the CLI's stability under both golden paths and edge cases.

---

## The BET Testing Methodology

Each package/component should aim to implement all three components of the BET framework:

1. **B - Benchmarks**: Measure performance, memory allocations, and CPU overhead for core functions (e.g., traversing directories, styling text, parsing GitLab rules, or reading configuration values).
2. **E - Examples**: Go executable examples (`ExampleFuncName`) to document correct UI output (e.g., logging formats and error banners) and serve as compile-time checked documentation.
3. **T - Tests**: Extensive unit, table-driven, and integration tests to verify functionality under different environment and config states.

---

## Execution Commands

### Run all tests

To run all tests in the codebase with coverage enabled:

```bash
CGO_ENABLED=1 go test ./... -cover -race -coverprofile=coverage.out -covermode=atomic
```

### Run Benchmarks

To run performance benchmarks and print allocation statistics:

```bash
go test -bench=. -benchmem ./...
```

### View Test Coverage

To view the generated coverage report in your browser:

```bash
go tool cover -html=coverage.out
```

---

## Mocking Strategies

To prevent mutating the host developer machine during test suite execution, the following testing patterns are standardized:

### 1. Mocking External CLI Dependencies (`exec.Command`)

Any package invoking external commands (e.g., `git`, `glab`, `tailscale`, `bw`) must define a package-level variable for command invocation rather than calling `exec.Command` directly:

```go
// Inside package file (e.g., ts.go or gitlab.go)
var execCommand = exec.Command
```

In your unit tests (`_test.go`), backup and override the `execCommand` variable to capture inputs and return controlled exit codes or mock outputs:

```go
// Inside test file
func TestCommand(t *testing.T) {
    original := execCommand
    defer func() { execCommand = original }()

    execCommand = func(name string, arg ...string) *exec.Cmd {
        // Return a mock executable command
        return exec.Command("echo", "mocked output")
    }
    
    // Call the command runner logic...
}
```

### 2. Mocking Interactive Prompts (`ui` Package)

To headless-test configuration or onboarding subcommands without hanging on interactive TUI prompts, mock the package-level prompt variables defined in `internal/ui`:

```go
import "github.com/eng618/eng/internal/ui"

func TestInteractiveCommand(t *testing.T) {
    oldConfirm := ui.Confirm
    oldInput := ui.Input
    defer func() {
        ui.Confirm = oldConfirm
        ui.Input = oldInput
    }()

    ui.Confirm = func(prompt string, defaultVal bool) (bool, error) {
        return true, nil
    }
    ui.Input = func(prompt, defaultVal string) (string, error) {
        return "mocked-value", nil
    }
    
    // Execute command...
}
```

### 3. Capturing CLI Output and Logging

When testing print operations, capture outputs by using the `log` package's redirectable writers:

```go
var outBuf bytes.Buffer
log.SetWriters(&outBuf, &outBuf)
defer log.ResetWriters()

// Run styled message functions
log.Success("Task complete")

// Assert output
if !strings.Contains(outBuf.String(), "Task complete") {
    t.Errorf("expected output to contain message")
}
```
