package ui

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/eng618/eng/internal/ui/theme"
)

var (
	sectionHeaderStyle = lipgloss.NewStyle().
				Foreground(theme.Primary).
				Bold(true).
				MarginTop(1).
				MarginBottom(0)

	commandNameStyle = lipgloss.NewStyle().
				Foreground(theme.Secondary).
				Bold(true)

	flagNameStyle = lipgloss.NewStyle().
			Foreground(theme.Primary)

	descriptionStyle = lipgloss.NewStyle().
				Foreground(theme.Foreground)

	mutedStyle = lipgloss.NewStyle().
			Foreground(theme.MutedForeground)

	usageSyntaxStyle = lipgloss.NewStyle().
				Foreground(theme.Foreground).
				Bold(true)
)

// HelpFunc returns a Cobra help function that formats help output using Lip Gloss styles.
func HelpFunc(c *cobra.Command, args []string) {
	out := c.OutOrStdout()
	renderHelp(out, c)
}

// UsageFunc returns a Cobra usage function that formats usage output using Lip Gloss styles.
func UsageFunc(c *cobra.Command) error {
	out := c.OutOrStderr()
	renderHelp(out, c)
	return nil
}

func renderHelp(w io.Writer, c *cobra.Command) {
	var b strings.Builder

	// Header / Description
	if c.Long != "" {
		b.WriteString(theme.BaseText.Render(strings.TrimSpace(c.Long)))
		b.WriteString("\n\n")
	} else if c.Short != "" {
		b.WriteString(theme.BaseText.Render(c.Short))
		b.WriteString("\n\n")
	}

	// Usage Syntax
	b.WriteString(sectionHeaderStyle.Render("USAGE"))
	b.WriteString("\n  ")
	if c.HasAvailableSubCommands() {
		b.WriteString(usageSyntaxStyle.Render(c.CommandPath() + " [command] [flags]"))
	} else {
		b.WriteString(usageSyntaxStyle.Render(c.UseLine()))
	}
	b.WriteString("\n")

	// Command Groups / Subcommands
	if c.HasAvailableSubCommands() {
		groups := c.Groups()
		availableCmds := c.Commands()

		if len(groups) > 0 {
			for _, group := range groups {
				var groupCmds []*cobra.Command
				for _, subCmd := range availableCmds {
					if subCmd.GroupID == group.ID && subCmd.IsAvailableCommand() {
						groupCmds = append(groupCmds, subCmd)
					}
				}
				if len(groupCmds) > 0 {
					b.WriteString("\n")
					b.WriteString(sectionHeaderStyle.Render(strings.ToUpper(group.Title)))
					b.WriteString("\n")
					renderCommandList(&b, groupCmds)
				}
			}

			// Render ungroupped available commands
			var ungrouppedCmds []*cobra.Command
			for _, subCmd := range availableCmds {
				if subCmd.GroupID == "" && subCmd.IsAvailableCommand() && subCmd.Name() != "help" {
					ungrouppedCmds = append(ungrouppedCmds, subCmd)
				}
			}
			if len(ungrouppedCmds) > 0 {
				b.WriteString("\n")
				b.WriteString(sectionHeaderStyle.Render("ADDITIONAL COMMANDS"))
				b.WriteString("\n")
				renderCommandList(&b, ungrouppedCmds)
			}
		} else {
			var validCmds []*cobra.Command
			for _, subCmd := range availableCmds {
				if subCmd.IsAvailableCommand() {
					validCmds = append(validCmds, subCmd)
				}
			}
			if len(validCmds) > 0 {
				b.WriteString("\n")
				b.WriteString(sectionHeaderStyle.Render("AVAILABLE COMMANDS"))
				b.WriteString("\n")
				renderCommandList(&b, validCmds)
			}
		}
	}

	// Local Flags
	localFlags := c.LocalFlags()
	if localFlags.HasAvailableFlags() {
		b.WriteString("\n")
		b.WriteString(sectionHeaderStyle.Render("FLAGS"))
		b.WriteString("\n")
		renderFlags(&b, localFlags)
	}

	// Persistent Flags
	inheritedFlags := c.InheritedFlags()
	if inheritedFlags.HasAvailableFlags() {
		b.WriteString("\n")
		b.WriteString(sectionHeaderStyle.Render("GLOBAL FLAGS"))
		b.WriteString("\n")
		renderFlags(&b, inheritedFlags)
	}

	// Examples
	if c.HasExample() {
		b.WriteString("\n")
		b.WriteString(sectionHeaderStyle.Render("EXAMPLES"))
		b.WriteString("\n")
		lines := strings.Split(c.Example, "\n")
		for _, line := range lines {
			if strings.TrimSpace(line) != "" {
				b.WriteString("  " + mutedStyle.Render(line) + "\n")
			}
		}
	}

	// Footer note if subcommands exist
	if c.HasAvailableSubCommands() {
		b.WriteString("\n")
		b.WriteString(
			mutedStyle.Render(
				fmt.Sprintf("Use \"%s [command] --help\" for more information about a command.", c.CommandPath()),
			),
		)
		b.WriteString("\n")
	}

	fmt.Fprint(w, b.String())
}

func renderCommandList(b *strings.Builder, cmds []*cobra.Command) {
	maxLen := 0
	for _, cmd := range cmds {
		if len(cmd.Name()) > maxLen {
			maxLen = len(cmd.Name())
		}
	}

	for _, cmd := range cmds {
		padding := strings.Repeat(" ", maxLen-len(cmd.Name())+4)
		nameStr := commandNameStyle.Render(cmd.Name())
		descStr := descriptionStyle.Render(cmd.Short)
		fmt.Fprintf(b, "  %s%s%s\n", nameStr, padding, descStr)
	}
}

func renderFlags(b *strings.Builder, flagSet *pflag.FlagSet) {
	var flags []*pflag.Flag
	maxLen := 0

	flagSet.VisitAll(func(f *pflag.Flag) {
		if f.Hidden {
			return
		}
		flags = append(flags, f)
		flagStr := formatFlagNames(f)
		if len(flagStr) > maxLen {
			maxLen = len(flagStr)
		}
	})

	for _, f := range flags {
		flagStr := formatFlagNames(f)
		padding := strings.Repeat(" ", maxLen-len(flagStr)+4)
		styledFlag := flagNameStyle.Render(flagStr)
		styledDesc := descriptionStyle.Render(f.Usage)

		if f.DefValue != "" && f.DefValue != "false" && f.DefValue != "[]" {
			styledDesc += mutedStyle.Render(fmt.Sprintf(" (default %q)", f.DefValue))
		}
		fmt.Fprintf(b, "  %s%s%s\n", styledFlag, padding, styledDesc)
	}
}

func formatFlagNames(f *pflag.Flag) string {
	if f.Shorthand != "" {
		return fmt.Sprintf("-%s, --%s", f.Shorthand, f.Name)
	}
	return fmt.Sprintf("    --%s", f.Name)
}
