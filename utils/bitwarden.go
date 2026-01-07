package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/eng618/eng/utils/log"
)

// BitwardenItem represents a Bitwarden vault item
type BitwardenItem struct {
	ID     string           `json:"id"`
	Name   string           `json:"name"`
	Fields []BitwardenField `json:"fields,omitempty"`
	Login  *BitwardenLogin  `json:"login,omitempty"`
	Notes  string           `json:"notes,omitempty"`
}

// BitwardenField represents a custom field in a Bitwarden item
type BitwardenField struct {
	Name  string `json:"name"`
	Value string `json:"value"`
	Type  int    `json:"type"`
}

// BitwardenLogin represents login credentials
type BitwardenLogin struct {
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}

// CheckBitwardenLoginStatus checks if the user is logged into Bitwarden
func CheckBitwardenLoginStatus() (bool, error) {
	cmd := exec.Command("bw", "status")
	output, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("failed to check Bitwarden status: %w", err)
	}

	var status struct {
		Status string `json:"status"`
	}
	if err := json.Unmarshal(output, &status); err != nil {
		return false, fmt.Errorf("failed to parse Bitwarden status: %w", err)
	}

	return status.Status == "unlocked", nil
}

// UnlockBitwardenVault prompts the user to unlock their Bitwarden vault
func UnlockBitwardenVault() error {
	log.Message("Bitwarden vault is locked. Please unlock it to retrieve SSH keys.")
	log.Message("Run: bw unlock")
	log.Message("Then set the session: export BW_SESSION='your-session-key'")
	log.Message("")
	log.Message("Alternatively, you can log in directly:")
	log.Message("bw login")
	log.Message("")

	return fmt.Errorf("Bitwarden vault is locked - please unlock and try again")
}

// EnsureBitwardenSession ensures the Bitwarden vault is unlocked and returns a session key.
// It first checks BW_SESSION and status; if locked, it prompts for the master password using
// `bw unlock --raw` and returns the session token. If unauthenticated, it attempts `bw login`.
func EnsureBitwardenSession() (string, error) {
	// If BW_SESSION already set and status is unlocked, reuse it
	if os.Getenv("BW_SESSION") != "" {
		unlocked, _ := CheckBitwardenLoginStatus()
		if unlocked {
			return os.Getenv("BW_SESSION"), nil
		}
	}
	// Check current status
	cmd := exec.Command("bw", "status")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("bitwarden status failed: %w", err)
	}
	var status struct {
		Status string `json:"status"`
	}
	if err := json.Unmarshal(out, &status); err != nil {
		return "", fmt.Errorf("parse bitwarden status failed: %w", err)
	}
	switch status.Status {
	case "unauthenticated":
		log.Info("Logging into Bitwarden (you may be prompted)")
		login := exec.Command("bw", "login")
		login.Stdin = os.Stdin
		login.Stdout = log.Writer()
		login.Stderr = log.ErrorWriter()
		if err := login.Run(); err != nil {
			return "", fmt.Errorf("bitwarden login failed: %w", err)
		}
		// fallthrough to unlock
		fallthrough
	case "locked":
		log.Info("Unlocking Bitwarden vault (enter master password)")
		unlock := exec.Command("bw", "unlock", "--raw")
		unlock.Stdin = os.Stdin
		tokenBytes, err := unlock.Output()
		if err != nil {
			return "", fmt.Errorf("bitwarden unlock failed: %w", err)
		}
		token := strings.TrimSpace(string(tokenBytes))
		if token == "" {
			return "", fmt.Errorf("empty BW_SESSION returned from unlock")
		}
		return token, nil
	case "unlocked":
		return os.Getenv("BW_SESSION"), nil
	default:
		return "", fmt.Errorf("unknown bitwarden status: %s", status.Status)
	}
}

// SaveOrUpdateBitwardenSecret saves a secret value into Bitwarden under the given item name.
// If the item exists, it will be updated; otherwise it is created. A custom field named
// "eng-cli" with value "true" is added to tag usage by this CLI.
func SaveOrUpdateBitwardenSecret(name, secret, notes string) (string, error) {
	// Build item JSON
	item := map[string]any{
		"type":  1, // login
		"name":  name,
		"notes": notes,
		"login": map[string]any{
			"password": secret,
		},
		"fields": []map[string]any{
			{"name": "eng-cli", "value": "true", "type": 0},
		},
	}
	b, err := json.Marshal(item)
	if err != nil {
		return "", err
	}
	// Ensure session
	sess, err := EnsureBitwardenSession()
	if err != nil {
		return "", err
	}
	env := append(os.Environ(), "BW_SESSION="+sess)

	// Try to find existing item
	find := exec.Command("bw", "list", "items", "--search", name)
	find.Env = env
	fb, err := find.Output()
	if err != nil {
		return "", fmt.Errorf("bitwarden list items failed: %w", err)
	}
	var existing []BitwardenItem
	_ = json.Unmarshal(fb, &existing)

	// Encode JSON for bw create/edit
	enc := exec.Command("bw", "encode")
	enc.Env = env
	enc.Stdin = strings.NewReader(string(b))
	encoded, err := enc.Output()
	if err != nil {
		return "", fmt.Errorf("bitwarden encode failed: %w", err)
	}

	if len(existing) > 0 {
		id := existing[0].ID
		edit := exec.Command("bw", "edit", "item", id, string(encoded))
		edit.Env = env
		if _, err := edit.Output(); err != nil {
			return "", fmt.Errorf("bitwarden edit item failed: %w", err)
		}
		return id, nil
	}
	create := exec.Command("bw", "create", "item", string(encoded))
	create.Env = env
	out, err := create.Output()
	if err != nil {
		return "", fmt.Errorf("bitwarden create item failed: %w", err)
	}
	var created BitwardenItem
	if err := json.Unmarshal(out, &created); err != nil {
		return "", err
	}
	return created.ID, nil
}

// GetBitwardenItem retrieves an item from Bitwarden vault by name
func GetBitwardenItem(name string) (*BitwardenItem, error) {
	cmd := exec.Command("bw", "get", "item", name)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get item '%s' from Bitwarden: %w", name, err)
	}

	var item BitwardenItem
	if err := json.Unmarshal(output, &item); err != nil {
		return nil, fmt.Errorf("failed to parse Bitwarden item: %w", err)
	}

	return &item, nil
}

// ListBitwardenItems returns all items in the vault (filtered by type if specified)
func ListBitwardenItems() ([]BitwardenItem, error) {
	cmd := exec.Command("bw", "list", "items")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list Bitwarden items: %w", err)
	}

	var items []BitwardenItem
	if err := json.Unmarshal(output, &items); err != nil {
		return nil, fmt.Errorf("failed to parse Bitwarden items: %w", err)
	}

	return items, nil
}

// FindSSHKeysInVault searches for SSH key items in the Bitwarden vault
func FindSSHKeysInVault() ([]BitwardenItem, error) {
	items, err := ListBitwardenItems()
	if err != nil {
		return nil, err
	}

	var sshKeys []BitwardenItem
	for _, item := range items {
		// Look for items with "ssh" or "SSH" in the name
		if strings.Contains(strings.ToLower(item.Name), "ssh") {
			// Check if it contains SSH key data
			if hasSSHKeyData(&item) {
				sshKeys = append(sshKeys, item)
			}
		}
	}

	return sshKeys, nil
}

// hasSSHKeyData checks if a Bitwarden item contains SSH key data
func hasSSHKeyData(item *BitwardenItem) bool {
	// Check notes for SSH key format
	if strings.Contains(item.Notes, "-----BEGIN") && strings.Contains(item.Notes, "-----END") {
		return true
	}

	// Check custom fields for SSH keys
	for _, field := range item.Fields {
		if strings.Contains(strings.ToLower(field.Name), "ssh") ||
			strings.Contains(strings.ToLower(field.Name), "key") {
			if strings.Contains(field.Value, "-----BEGIN") && strings.Contains(field.Value, "-----END") {
				return true
			}
		}
	}

	// Check login password field for SSH keys
	if item.Login != nil && item.Login.Password != "" {
		if strings.Contains(item.Login.Password, "-----BEGIN") && strings.Contains(item.Login.Password, "-----END") {
			return true
		}
	}

	return false
}

// ExtractSSHKeyFromItem extracts SSH private key from a Bitwarden item
func ExtractSSHKeyFromItem(item *BitwardenItem) (string, error) {
	// First check notes
	if strings.Contains(item.Notes, "-----BEGIN") && strings.Contains(item.Notes, "-----END") {
		return item.Notes, nil
	}

	// Check custom fields
	for _, field := range item.Fields {
		if strings.Contains(strings.ToLower(field.Name), "private") ||
			strings.Contains(strings.ToLower(field.Name), "key") {
			if strings.Contains(field.Value, "-----BEGIN") && strings.Contains(field.Value, "-----END") {
				return field.Value, nil
			}
		}
	}

	// Check login password as last resort
	if item.Login != nil && item.Login.Password != "" {
		if strings.Contains(item.Login.Password, "-----BEGIN") && strings.Contains(item.Login.Password, "-----END") {
			return item.Login.Password, nil
		}
	}

	return "", fmt.Errorf("no SSH private key found in item '%s'", item.Name)
}
