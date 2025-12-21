package utils

import (
	"encoding/json"
	"fmt"
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
