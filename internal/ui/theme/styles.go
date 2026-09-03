package theme

import (
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/charmbracelet/lipgloss"
)

// outMu guards Out/Err against concurrent use (tests run with -race).
var outMu sync.RWMutex

// Out and Err are the writers used for styled banner messages.
// Out is for success/info/warning banners (normal output),
// Err is for error banners. Tests may replace them via SetWriters.
// log.SetWriters propagates to these to keep output unified.
var (
	Out io.Writer = os.Stdout
	Err io.Writer = os.Stderr
)

// SetWriters replaces Out/Err in a thread-safe way.
func SetWriters(out, errOut io.Writer) {
	outMu.Lock()
	defer outMu.Unlock()
	if out != nil {
		Out = out
	}
	if errOut != nil {
		Err = errOut
	}
}

func getOut() io.Writer {
	outMu.RLock()
	defer outMu.RUnlock()
	return Out
}

func getErr() io.Writer {
	outMu.RLock()
	defer outMu.RUnlock()
	return Err
}

var (
	// BaseText is the default text style
	BaseText = lipgloss.NewStyle().Foreground(Foreground)

	// MutedText is used for secondary information
	MutedText = lipgloss.NewStyle().Foreground(MutedForeground)

	// BoldText highlights text in the default foreground
	BoldText = BaseText.Bold(true)

	// PrimaryText is text colored in the primary brand color
	PrimaryText = lipgloss.NewStyle().Foreground(Primary)

	// SuccessText is text colored in the secondary/success color
	SuccessText = lipgloss.NewStyle().Foreground(Secondary)

	// ErrorText is text colored in the destructive color
	ErrorText = lipgloss.NewStyle().Foreground(Destructive).Bold(true)

	// InfoBox is a generic callout box with a primary border
	InfoBox = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(Primary).
		Padding(0, 1).
		MarginBottom(1)

	// SuccessBanner is a highlighted block for success messages
	SuccessBanner = lipgloss.NewStyle().
			Background(Secondary).
			Foreground(Background).
			Bold(true).
			Padding(0, 1).
			MarginRight(1)

	// WarningBanner is a highlighted block for warnings
	WarningBanner = lipgloss.NewStyle().
			Background(lipgloss.AdaptiveColor{Light: "#f59e0b", Dark: "#d97706"}). // Yellow/Orange
			Foreground(Background).
			Bold(true).
			Padding(0, 1).
			MarginRight(1)

	// ErrorBanner is a highlighted block for critical errors
	ErrorBanner = lipgloss.NewStyle().
			Background(Destructive).
			Foreground(Background).
			Bold(true).
			Padding(0, 1).
			MarginRight(1)
)

// SuccessMessage prints a formatted success message.
func SuccessMessage(msg string) {
	banner := SuccessBanner.Render("SUCCESS")
	text := BaseText.Render(msg)
	fmt.Fprintf(getOut(), "%s %s\n", banner, text)
}

// InfoMessage prints a formatted informational message.
func InfoMessage(msg string) {
	banner := lipgloss.NewStyle().
		Background(Primary).
		Foreground(Background).
		Bold(true).
		Padding(0, 1).
		MarginRight(1).
		Render("INFO")
	text := BaseText.Render(msg)
	fmt.Fprintf(getOut(), "%s %s\n", banner, text)
}

// ErrorMessage prints a formatted error message.
func ErrorMessage(msg string) {
	banner := ErrorBanner.Render("ERROR")
	text := ErrorText.Render(msg)
	fmt.Fprintf(getErr(), "%s %s\n", banner, text)
}

// WarningMessage prints a formatted warning message.
func WarningMessage(msg string) {
	banner := WarningBanner.Render("WARN")
	text := BaseText.Render(msg)
	fmt.Fprintf(getOut(), "%s %s\n", banner, text)
}
