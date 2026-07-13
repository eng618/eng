package theme

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// ActionableError is an interface that extends the standard error interface
// with a Suggestion() method. Implement this to provide helpful tips for resolving errors.
type ActionableError interface {
	error
	Suggestion() string
}

// actionableErrorImpl is the internal implementation of ActionableError.
type actionableErrorImpl struct {
	err        error
	suggestion string
}

func (e *actionableErrorImpl) Error() string {
	return e.err.Error()
}

func (e *actionableErrorImpl) Suggestion() string {
	return e.suggestion
}

func (e *actionableErrorImpl) Unwrap() error {
	return e.err
}

// NewActionableError wraps an existing error with a helpful suggestion.
func NewActionableError(err error, suggestion string) error {
	if err == nil {
		return nil
	}
	return &actionableErrorImpl{
		err:        err,
		suggestion: suggestion,
	}
}

// ErrorBoxStyle defines the container style for the rich error box.
var ErrorBoxStyle = lipgloss.NewStyle().
	Border(lipgloss.RoundedBorder()).
	BorderForeground(Destructive).
	Padding(1, 2).
	MarginTop(1).
	MarginBottom(1)

// HandleError takes an error, styles it into a rich CLI box, prints it to stderr, and exits.
// If the error implements ActionableError, it appends the suggestion.
func HandleError(err error) {
	if err == nil {
		return
	}

	var content strings.Builder

	// The Error Title
	content.WriteString(ErrorBanner.Render("ERROR") + " " + ErrorText.Render(err.Error()))

	// Check if it's an ActionableError
	var actErr ActionableError
	if errors.As(err, &actErr) {
		content.WriteString("\n\n")
		content.WriteString(PrimaryText.Bold(true).Render("💡 Suggestion: "))
		content.WriteString(BaseText.Render(actErr.Suggestion()))
	}

	// Print the styled box
	fmt.Fprintln(os.Stderr, ErrorBoxStyle.Render(content.String()))
}
