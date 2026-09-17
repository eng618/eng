package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// CurrentVersion is the latest config schema version stored as the
// top-level `version` key. Files without a version are treated as v0.
const CurrentVersion = 2

// Schema renames applied by MigrateMap (old dotted key -> new dotted key).
// v1 renames predate versioning; v2 normalizes email/gitlab keys and drops
// the dotfiles.local_repo_path alias.
var schemaRenames = []struct{ old, new string }{
	{"dotfiles.repoPath", "dotfiles.bare_repo_path"},
	{"dotfiles.worktree", "dotfiles.worktree_path"},
	{"dotfiles.workTree", "dotfiles.worktree_path"},
	{"dotfiles.local_repo_path", "dotfiles.bare_repo_path"},
	{"git.devPath", "git.dev_path"},
	{"user-email", "email"},
	{"gitlab.tokenItem", "gitlab.token_item"},
}

// MigrateMap migrates a decoded config map to the current schema version.
// It renames legacy keys (deleting the old key only when the new key is
// empty), ensures `projects` exists, and stamps the current version.
// It returns the migrated map and whether anything changed.
func MigrateMap(in map[string]any) (map[string]any, bool, error) {
	if in == nil {
		in = map[string]any{}
	}
	changed := false
	for _, r := range schemaRenames {
		oldVal, oldOk := getDotted(in, r.old)
		if !oldOk || isEmpty(oldVal) {
			continue
		}
		if newVal, newOk := getDotted(in, r.new); newOk && !isEmpty(newVal) {
			// New key wins; drop the legacy key.
			deleteDotted(in, r.old)
			changed = true
			continue
		}
		setDotted(in, r.new, oldVal)
		deleteDotted(in, r.old)
		changed = true
	}
	if _, ok := in["projects"]; !ok {
		in["projects"] = []any{}
		changed = true
	}
	if v, _ := in["version"].(int); v != CurrentVersion {
		in["version"] = CurrentVersion
		changed = true
	}
	return in, changed, nil
}

// BackupFile copies path to path.bak.<UTC timestamp> and returns the backup path.
func BackupFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read config for backup: %w", err)
	}
	backup := fmt.Sprintf("%s.bak.%s", path, time.Now().UTC().Format("20060102T150405Z"))
	if err := os.WriteFile(backup, data, 0o600); err != nil {
		return "", fmt.Errorf("write config backup: %w", err)
	}
	return backup, nil
}

// MigrateFile loads a YAML config file, migrates it to the current schema,
// writes a timestamped backup when changes are needed, and rewrites the file.
// It is idempotent: a second run reports changed=false.
func MigrateFile(path string) (backup string, changed bool, err error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", false, fmt.Errorf("read config: %w", err)
	}
	var decoded map[string]any
	if err := yaml.Unmarshal(data, &decoded); err != nil {
		return "", false, fmt.Errorf("parse config: %w", err)
	}
	if decoded == nil {
		decoded = map[string]any{}
	}
	migrated, changed, err := MigrateMap(decoded)
	if err != nil {
		return "", false, err
	}
	if !changed {
		return "", false, nil
	}
	backup, err = BackupFile(path)
	if err != nil {
		return "", false, err
	}
	out, err := yaml.Marshal(migrated)
	if err != nil {
		return backup, true, fmt.Errorf("encode config: %w", err)
	}
	if err := os.WriteFile(path, out, 0o600); err != nil {
		return backup, true, fmt.Errorf("write migrated config: %w", err)
	}
	return backup, true, nil
}

func isEmpty(v any) bool {
	switch t := v.(type) {
	case nil:
		return true
	case string:
		return t == ""
	default:
		return false
	}
}

func getDotted(m map[string]any, dotted string) (any, bool) {
	var cur any = m
	for _, part := range splitDotted(dotted) {
		cm, ok := cur.(map[string]any)
		if !ok {
			return nil, false
		}
		cur, ok = cm[part]
		if !ok {
			return nil, false
		}
	}
	return cur, true
}

func setDotted(m map[string]any, dotted string, val any) {
	parts := splitDotted(dotted)
	cur := m
	for _, p := range parts[:len(parts)-1] {
		next, ok := cur[p].(map[string]any)
		if !ok {
			next = map[string]any{}
			cur[p] = next
		}
		cur = next
	}
	cur[parts[len(parts)-1]] = val
}

func deleteDotted(m map[string]any, dotted string) {
	parts := splitDotted(dotted)
	cur := m
	for _, p := range parts[:len(parts)-1] {
		next, ok := cur[p].(map[string]any)
		if !ok {
			return
		}
		cur = next
	}
	delete(cur, parts[len(parts)-1])
}

func splitDotted(s string) []string {
	var parts []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '.' {
			parts = append(parts, s[start:i])
			start = i + 1
		}
	}
	return append(parts, s[start:])
}
