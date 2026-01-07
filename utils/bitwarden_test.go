package utils

import (
	"testing"
)

func TestHasSSHKeyData(t *testing.T) {
	tests := []struct {
		name     string
		item     BitwardenItem
		expected bool
	}{
		{
			name: "SSH key in notes",
			item: BitwardenItem{
				Notes: "-----BEGIN OPENSSH PRIVATE KEY-----\nkey data\n-----END OPENSSH PRIVATE KEY-----",
			},
			expected: true,
		},
		{
			name: "SSH key in custom field",
			item: BitwardenItem{
				Fields: []BitwardenField{
					{
						Name:  "SSH Private Key",
						Value: "-----BEGIN OPENSSH PRIVATE KEY-----\nkey data\n-----END OPENSSH PRIVATE KEY-----",
						Type:  0,
					},
				},
			},
			expected: true,
		},
		{
			name: "SSH key in password field",
			item: BitwardenItem{
				Login: &BitwardenLogin{
					Password: "-----BEGIN OPENSSH PRIVATE KEY-----\nkey data\n-----END OPENSSH PRIVATE KEY-----",
				},
			},
			expected: true,
		},
		{
			name: "No SSH key data",
			item: BitwardenItem{
				Name: "Regular Password",
				Login: &BitwardenLogin{
					Username: "user",
					Password: "password123",
				},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := hasSSHKeyData(&tt.item)
			if result != tt.expected {
				t.Errorf("hasSSHKeyData() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestExtractSSHKeyFromItem(t *testing.T) {
	tests := []struct {
		name        string
		item        BitwardenItem
		expectedKey string
		expectError bool
	}{
		{
			name: "Extract from notes",
			item: BitwardenItem{
				Name:  "SSH Key",
				Notes: "-----BEGIN OPENSSH PRIVATE KEY-----\nkey data\n-----END OPENSSH PRIVATE KEY-----",
			},
			expectedKey: "-----BEGIN OPENSSH PRIVATE KEY-----\nkey data\n-----END OPENSSH PRIVATE KEY-----",
			expectError: false,
		},
		{
			name: "Extract from custom field",
			item: BitwardenItem{
				Name: "SSH Key",
				Fields: []BitwardenField{
					{
						Name:  "Private Key",
						Value: "-----BEGIN OPENSSH PRIVATE KEY-----\nkey data\n-----END OPENSSH PRIVATE KEY-----",
						Type:  0,
					},
				},
			},
			expectedKey: "-----BEGIN OPENSSH PRIVATE KEY-----\nkey data\n-----END OPENSSH PRIVATE KEY-----",
			expectError: false,
		},
		{
			name: "Extract from password field",
			item: BitwardenItem{
				Name: "SSH Key",
				Login: &BitwardenLogin{
					Password: "-----BEGIN OPENSSH PRIVATE KEY-----\nkey data\n-----END OPENSSH PRIVATE KEY-----",
				},
			},
			expectedKey: "-----BEGIN OPENSSH PRIVATE KEY-----\nkey data\n-----END OPENSSH PRIVATE KEY-----",
			expectError: false,
		},
		{
			name: "No SSH key found",
			item: BitwardenItem{
				Name: "Regular Item",
				Login: &BitwardenLogin{
					Username: "user",
					Password: "password123",
				},
			},
			expectedKey: "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key, err := ExtractSSHKeyFromItem(&tt.item)
			if tt.expectError {
				if err == nil {
					t.Errorf("ExtractSSHKeyFromItem() expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("ExtractSSHKeyFromItem() unexpected error: %v", err)
				}
				if key != tt.expectedKey {
					t.Errorf("ExtractSSHKeyFromItem() = %v, expected %v", key, tt.expectedKey)
				}
			}
		})
	}
}
