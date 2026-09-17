package config

import (
	"fmt"
	"net/mail"
	"net/url"
	"os"
	"strings"
)

// FieldError describes one invalid config field with a fix hint.
type FieldError struct {
	Field string
	Value string
	Hint  string
}

func (e FieldError) Error() string {
	if e.Hint != "" {
		return fmt.Sprintf("invalid %s %q: %s", e.Field, e.Value, e.Hint)
	}
	return fmt.Sprintf("invalid %s %q", e.Field, e.Value)
}

// Validate checks a ResolvedConfig and returns one FieldError per problem.
// Empty email/repo URL are allowed (prompt-on-empty covers first run);
// malformed values are reported with actionable hints.
func (c *ResolvedConfig) Validate() []FieldError {
	var errs []FieldError
	if c.Email != "" {
		if _, err := mail.ParseAddress(c.Email); err != nil {
			errs = append(errs, FieldError{
				Field: "email", Value: c.Email,
				Hint: "set with `eng config set email you@example.com`",
			})
		}
	}
	if c.DotfilesRepo != "" {
		if _, err := url.ParseRequestURI(c.DotfilesRepo); err != nil {
			if !isSCPStyle(c.DotfilesRepo) {
				errs = append(errs, FieldError{
					Field: "dotfiles.repo_url", Value: c.DotfilesRepo,
					Hint: "use an https:// or git@ (scp-style) URL",
				})
			}
		}
	}
	if c.DotfilesBranch != "" && strings.ContainsAny(c.DotfilesBranch, " \t~^:?*[") {
		errs = append(errs, FieldError{
			Field: "dotfiles.branch", Value: c.DotfilesBranch,
			Hint: "branch names cannot contain spaces or ~^:?*[",
		})
	}
	for field, path := range map[string]string{
		"git.dev_path":    c.GitDevPath,
		"containers.path": c.ContainersPath,
	} {
		if path == "" {
			continue
		}
		if err := checkDirField(field, path); err != nil {
			errs = append(errs, *err)
		}
	}
	return errs
}

// checkDirField ensures path exists and is a directory, or its parent exists
// so it can be created on first use.
func checkDirField(field, path string) *FieldError {
	info, err := os.Stat(path)
	if err == nil {
		if !info.IsDir() {
			return &FieldError{Field: field, Value: path, Hint: "path exists but is not a directory"}
		}
		return nil
	}
	if !os.IsNotExist(err) {
		return &FieldError{Field: field, Value: path, Hint: "cannot stat path: " + err.Error()}
	}
	return nil
}

func isSCPStyle(s string) bool {
	at := strings.Index(s, "@")
	colon := strings.Index(s, ":")
	return at > 0 && colon > at
}
