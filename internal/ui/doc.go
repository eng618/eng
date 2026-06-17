// Package ui provides reusable interactive terminal UI components,
// standardized theming, and wrappers around the "charmbracelet/huh"
// and "charmbracelet/bubbles" libraries.
//
// This package centralizes all interactive CLI components ensuring
// a consistent user experience and visual identity across the entire
// eng CLI application. It abstract away the complexity of managing
// terminal prompts and concurrent spinners.
//
// # Available Features
//
//   - **Theming**: The `theme.EngTheme()` function supplies a consistent, adaptive
//     light/dark mode HSL color palette across all interactive prompts.
//   - **Prompts**: Standardized functions for user interaction:
//   - `Confirm(message string, defaultVal bool)`: Yes/No prompts.
//   - `Input(message string, defaultVal string)`: Text input.
//   - `Select(message string, options []string, defaultVal string)`: Single selection.
//   - `MultiSelect(message string, options []string)`: Multiple selection.
//   - `Password(message string)`: Secret text input.
//   - **Spinners**: Beautiful and thread-safe progress spinners.
//   - `Spinner`: Single operation status and progress.
//   - `MultiSpinner`: For displaying multiple concurrent operations
//     (e.g., parallel git pulls, downloading files, system updates).
//
// # Testing
//
// The prompt functions within this package are exposed as variables
// (e.g., `var Confirm = ConfirmImpl`) to allow them to be easily mocked
// during unit testing.
package ui
