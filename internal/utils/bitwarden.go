package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/eng618/eng/internal/utils/log"
	"golang.org/x/term"
)

// BitwardenItem represents a Bitwarden vault item.
type BitwardenItem struct {
	ID     string           `json:"id"`
	Name   string           `json:"name"`
	Type   int              `json:"type,omitempty"`
	Fields []BitwardenField `json:"fields,omitempty"`
	Login  *BitwardenLogin  `json:"login,omitempty"`
	Notes  string           `json:"notes,omitempty"`
	SSHKey *BitwardenSSHKey `json:"sshKey,omitempty"`
}

// BitwardenSSHKey represents native SSH key payloads in Bitwarden items.
type BitwardenSSHKey struct {
	PrivateKey  string `json:"privateKey,omitempty"`
	PublicKey   string `json:"publicKey,omitempty"`
	Fingerprint string `json:"fingerprint,omitempty"`
}

// BitwardenField represents a custom field in a Bitwarden item.
type BitwardenField struct {
	Name  string `json:"name"`
	Value string `json:"value"`
	Type  int    `json:"type"`
}

// BitwardenLogin represents login credentials.
type BitwardenLogin struct {
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}

// CheckBitwardenLoginStatus checks if the user is logged into Bitwarden.
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

// UnlockBitwardenVault prompts the user to unlock their Bitwarden vault.
func UnlockBitwardenVault() error {
	log.Message("Bitwarden vault is locked. Please unlock it to retrieve SSH keys.")
	log.Message("Run: bw unlock")
	log.Message("Then set the session: export BW_SESSION='your-session-key'")
	log.Message("")
	log.Message("Alternatively, you can log in directly:")
	log.Message("bw login")
	log.Message("")

	return fmt.Errorf("bitwarden vault is locked - please unlock and try again")
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

		// Refresh status after login.
		cmd = exec.Command("bw", "status")
		out, err = cmd.Output()
		if err != nil {
			return "", fmt.Errorf("bitwarden status failed after login: %w", err)
		}
		if err := json.Unmarshal(out, &status); err != nil {
			return "", fmt.Errorf("parse bitwarden status failed after login: %w", err)
		}

		if status.Status == "unlocked" {
			session := os.Getenv("BW_SESSION")
			if session != "" {
				return session, nil
			}
		}

		if status.Status != "locked" {
			return "", fmt.Errorf("unexpected bitwarden status after login: %s", status.Status)
		}

		fallthrough
	case "locked":
		token, err := unlockBitwardenVault()
		if err != nil {
			return "", err
		}
		if err := os.Setenv("BW_SESSION", token); err != nil {
			return "", fmt.Errorf("failed to set BW_SESSION: %w", err)
		}
		return token, nil
	case "unlocked":
		session := os.Getenv("BW_SESSION")
		if session != "" {
			return session, nil
		}

		// If unlocked without BW_SESSION exported in this process, get a fresh session token.
		token, err := unlockBitwardenVault()
		if err != nil {
			return "", err
		}
		if err := os.Setenv("BW_SESSION", token); err != nil {
			return "", fmt.Errorf("failed to set BW_SESSION: %w", err)
		}
		return token, nil
	default:
		return "", fmt.Errorf("unknown bitwarden status: %s", status.Status)
	}
}

func unlockBitwardenVault() (string, error) {
	log.Info("Unlocking Bitwarden vault (enter master password)")

	passwordEnv := "BW_MASTER_PASSWORD_ENG"
	masterPassword := os.Getenv(passwordEnv)
	if masterPassword == "" {
		fmt.Fprint(os.Stderr, "Bitwarden master password: ")
		pw, err := term.ReadPassword(int(os.Stdin.Fd()))
		fmt.Fprintln(os.Stderr)
		if err != nil {
			return "", fmt.Errorf("failed to read bitwarden master password: %w", err)
		}
		masterPassword = strings.TrimSpace(string(pw))
		if masterPassword == "" {
			return "", fmt.Errorf("empty bitwarden master password")
		}
	}

	unlock := exec.Command("bw", "unlock", "--raw", "--passwordenv", passwordEnv)
	unlock.Env = append(os.Environ(), passwordEnv+"="+masterPassword)
	unlock.Stderr = log.ErrorWriter()
	tokenBytes, err := unlock.Output()
	if err != nil {
		return "", fmt.Errorf("bitwarden unlock failed: %w", err)
	}

	token := strings.TrimSpace(string(tokenBytes))
	if token == "" {
		return "", fmt.Errorf("empty BW_SESSION returned from unlock")
	}

	return token, nil
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

// GetBitwardenItem retrieves an item from Bitwarden vault by name.
func GetBitwardenItem(name string) (*BitwardenItem, error) {
	sess, err := EnsureBitwardenSession()
	if err != nil {
		return nil, err
	}

	cmd := exec.Command("bw", "get", "item", name)
	cmd.Env = append(os.Environ(), "BW_SESSION="+sess)
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

// ListBitwardenItems returns all items in the vault (filtered by type if specified).
func ListBitwardenItems() ([]BitwardenItem, error) {
	sess, err := EnsureBitwardenSession()
	if err != nil {
		return nil, err
	}

	cmd := exec.Command("bw", "list", "items")
	cmd.Env = append(os.Environ(), "BW_SESSION="+sess)
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

// FindSSHKeysInVault searches for SSH key items in the Bitwarden vault.
func FindSSHKeysInVault() ([]BitwardenItem, error) {
	items, err := ListBitwardenItems()
	if err != nil {
		return nil, err
	}

	var sshKeys []BitwardenItem
	for _, item := range items {
		name := strings.ToLower(item.Name)
		if strings.Contains(name, "ssh") || strings.Contains(name, "github") || item.Type == 5 {
			// Fetch full item details because `bw list items` may omit sensitive fields.
			fullItem, getErr := GetBitwardenItem(item.ID)
			if getErr != nil {
				log.Verbose(true, "Failed to load Bitwarden item %s: %v", item.ID, getErr)
				continue
			}

			if hasSSHKeyData(fullItem) {
				sshKeys = append(sshKeys, *fullItem)
			}
		}
	}

	return sshKeys, nil
}

// hasSSHKeyData checks if a Bitwarden item contains SSH key data.
func hasSSHKeyData(item *BitwardenItem) bool {
	// Native SSH key type payload.
	if item.SSHKey != nil && strings.TrimSpace(item.SSHKey.PrivateKey) != "" {
		return true
	}

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

// ExtractSSHKeyFromItem extracts SSH private key from a Bitwarden item.
func ExtractSSHKeyFromItem(item *BitwardenItem) (string, error) {
	if item.SSHKey != nil && strings.TrimSpace(item.SSHKey.PrivateKey) != "" {
		return item.SSHKey.PrivateKey, nil
	}

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
