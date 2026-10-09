package config

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/spf13/viper"

	"github.com/eng618/eng/internal/log"
	"github.com/eng618/eng/internal/paths"
	"github.com/eng618/eng/internal/ui/theme"
)

// ComposeStackEntry is a user-registered named compose stack pointing at an
// arbitrary local root path. The compose file itself is autodetected inside
// Path at discovery time (docker-compose.yml/yaml, compose.yml/yaml).
type ComposeStackEntry struct {
	Name string `mapstructure:"name" yaml:"name"`
	Path string `mapstructure:"path" yaml:"path"`
}

var composeStackNameRe = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_-]*$`)

// ValidateComposeStackName reports whether name is usable as a stack name.
func ValidateComposeStackName(name string) error {
	if !composeStackNameRe.MatchString(name) {
		return fmt.Errorf(
			"invalid stack name %q: use letters, digits, '-' or '_' starting with alphanumeric",
			name,
		)
	}
	return nil
}

// GetComposeStacks retrieves the registered compose stacks from the config file.
func GetComposeStacks() []ComposeStackEntry {
	var stacks []ComposeStackEntry
	if !viper.IsSet("containers.stacks") {
		return []ComposeStackEntry{}
	}
	if err := viper.UnmarshalKey("containers.stacks", &stacks); err != nil {
		log.Error("Failed to unmarshal containers.stacks configuration: %v", err)
		return []ComposeStackEntry{}
	}
	if stacks == nil {
		return []ComposeStackEntry{}
	}
	return stacks
}

// SaveComposeStacks persists the registered compose stacks to the config file.
func SaveComposeStacks(stacks []ComposeStackEntry) error {
	if stacks == nil {
		stacks = []ComposeStackEntry{}
	}
	viper.Set("containers.stacks", stacks)
	if err := viper.WriteConfig(); err != nil {
		return fmt.Errorf("failed to save compose stacks configuration: %w", err)
	}
	return nil
}

// AddComposeStack adds a new named stack or updates the path of an existing
// one (case-insensitive match). Path is stored as given; expansion happens on read.
func AddComposeStack(entry ComposeStackEntry) error {
	if err := ValidateComposeStackName(entry.Name); err != nil {
		return err
	}
	if strings.TrimSpace(entry.Path) == "" {
		return fmt.Errorf("stack path must not be empty")
	}
	stacks := GetComposeStacks()
	for i := range stacks {
		if strings.EqualFold(stacks[i].Name, entry.Name) {
			stacks[i] = entry
			return SaveComposeStacks(stacks)
		}
	}
	return SaveComposeStacks(append(stacks, entry))
}

// RemoveComposeStack removes a registered stack by name (case-insensitive).
func RemoveComposeStack(name string) error {
	stacks := GetComposeStacks()
	kept := make([]ComposeStackEntry, 0, len(stacks))
	found := false
	for _, s := range stacks {
		if strings.EqualFold(s.Name, name) {
			found = true
			continue
		}
		kept = append(kept, s)
	}
	if !found {
		return fmt.Errorf("compose stack %q not found", name)
	}
	return SaveComposeStacks(kept)
}

// ExpandedComposeStacks returns registered stacks with paths expanded.
func ExpandedComposeStacks() []ComposeStackEntry {
	stacks := GetComposeStacks()
	for i := range stacks {
		stacks[i].Path = paths.Expand(stacks[i].Path)
	}
	return stacks
}

// PromptComposeStackValuesFunc collects a stack name and root path interactively.
type PromptComposeStackValuesFunc func(initialName, initialPath string) (name, path string, err error)

// PromptComposeStackValuesImpl uses huh.NewForm to collect stack details interactively.
func PromptComposeStackValuesImpl(initialName, initialPath string) (name, path string, err error) {
	name = initialName
	path = initialPath

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Stack Name").
				Description("A unique name for this compose stack (e.g. media, homelab).").
				Value(&name).
				Validate(func(s string) error {
					return ValidateComposeStackName(strings.TrimSpace(s))
				}),
			huh.NewInput().
				Title("Stack Path").
				Description("Local directory containing the compose file (autodetected inside).").
				Value(&path).
				Validate(func(s string) error {
					return validateComposeStackPath(s)
				}),
		),
	).WithTheme(theme.EngTheme())

	err = form.Run()
	return strings.TrimSpace(name), strings.TrimSpace(path), err
}

// PromptComposeStackValues prompts for stack values; override in tests.
var PromptComposeStackValues = PromptComposeStackValuesImpl

// SelectComposeStack prompts to pick one registered stack and returns its index.
func SelectComposeStack(stacks []ComposeStackEntry) (int, error) {
	if len(stacks) == 0 {
		return -1, errors.New("no compose stacks registered")
	}

	options := make([]string, len(stacks))
	for i, s := range stacks {
		options[i] = s.Name
	}

	selected, err := SelectPrompt("Select a compose stack:", options, "")
	if err != nil {
		return -1, err
	}

	for i, opt := range options {
		if opt == selected {
			return i, nil
		}
	}
	return -1, errors.New("no compose stack selected")
}

// validateComposeStackPath ensures the path is a non-empty existing directory.
// Compose-file presence is checked by the caller (containers autodetection).
func validateComposeStackPath(path string) error {
	trimmed := strings.TrimSpace(path)
	if trimmed == "" {
		return errors.New("stack path is required")
	}
	info, err := os.Stat(paths.Expand(trimmed))
	if err != nil {
		return fmt.Errorf("cannot access path: %w", err)
	}
	if !info.IsDir() {
		return errors.New("stack path is not a directory")
	}
	return nil
}
