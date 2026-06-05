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
			name: "Native SSH key payload",
			item: BitwardenItem{
				SSHKey: &BitwardenSSHKey{
					PrivateKey: "-----BEGIN OPENSSH PRIVATE KEY-----\nkey data\n-----END OPENSSH PRIVATE KEY-----",
					PublicKey:  "ssh-ed25519 AAA...",
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
			name: "Extract from native SSH key payload",
			item: BitwardenItem{
				Name: "SSH Key",
				SSHKey: &BitwardenSSHKey{
					PrivateKey: "-----BEGIN OPENSSH PRIVATE KEY-----\nkey data\n-----END OPENSSH PRIVATE KEY-----",
				},
			},
			expectedKey: "-----BEGIN OPENSSH PRIVATE KEY-----\nkey data\n-----END OPENSSH PRIVATE KEY-----",
			expectError: false,
		},
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
		{
			name: "Native SSH key payload with whitespace PrivateKey should fall back to notes",
			item: BitwardenItem{
				Name: "SSH Key Empty",
				SSHKey: &BitwardenSSHKey{
					PrivateKey: "   \n\t  ",
				},
				Notes: "-----BEGIN OPENSSH PRIVATE KEY-----\nnotes key data\n-----END OPENSSH PRIVATE KEY-----",
			},
			expectedKey: "-----BEGIN OPENSSH PRIVATE KEY-----\nnotes key data\n-----END OPENSSH PRIVATE KEY-----",
			expectError: false,
		},
		{
			name: "Custom field with invalid name but valid format should be ignored",
			item: BitwardenItem{
				Name: "SSH Key Invalid Field",
				Fields: []BitwardenField{
					{
						Name:  "Random Field",
						Value: "-----BEGIN OPENSSH PRIVATE KEY-----\nrandom key data\n-----END OPENSSH PRIVATE KEY-----",
						Type:  0,
					},
				},
			},
			expectedKey: "",
			expectError: true,
		},
		{
			name: "Custom field with valid name but invalid format should be ignored",
			item: BitwardenItem{
				Name: "SSH Key Invalid Format Field",
				Fields: []BitwardenField{
					{
						Name:  "Private Key",
						Value: "just some random text without markers",
						Type:  0,
					},
				},
			},
			expectedKey: "",
			expectError: true,
		},
		{
			name: "Login password lacking proper format should fall through to error",
			item: BitwardenItem{
				Name: "Login password invalid format",
				Login: &BitwardenLogin{
					Password: "regularpassword",
				},
			},
			expectedKey: "",
			expectError: true,
		},
		{
			name: "Notes lacking proper format should fall through to next checks",
			item: BitwardenItem{
				Name:  "Notes invalid format, fallback to field",
				Notes: "just some notes without keys",
				Fields: []BitwardenField{
					{
						Name:  "private",
						Value: "-----BEGIN OPENSSH PRIVATE KEY-----\nfield key data\n-----END OPENSSH PRIVATE KEY-----",
						Type:  0,
					},
				},
			},
			expectedKey: "-----BEGIN OPENSSH PRIVATE KEY-----\nfield key data\n-----END OPENSSH PRIVATE KEY-----",
			expectError: false,
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
