package config

import (
	"testing"
)

// TestConfigValidate tests the Config.Validate method.
func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name        string
		permissions uint32
		wantErr     bool
		description string
	}{
		{
			name:        "ValidOwnerReadWrite",
			permissions: 0600,
			wantErr:     false,
			description: "Owner read/write only",
		},
		{
			name:        "ValidOwnerRWGroupOtherRead",
			permissions: 0644,
			wantErr:     false,
			description: "Owner read/write, group/other read",
		},
		{
			name:        "ValidOwnerRWXGroupOtherRX",
			permissions: 0755,
			wantErr:     false,
			description: "Owner read/write/execute, group/other read/execute",
		},
		{
			name:        "ValidOwnerRWGroupRead",
			permissions: 0640,
			wantErr:     false,
			description: "Owner read/write, group read",
		},
		{
			name:        "ValidOwnerOnly",
			permissions: 0400,
			wantErr:     false,
			description: "Owner read only",
		},
		{
			name:        "InvalidAllReadWrite",
			permissions: 0666,
			wantErr:     true,
			description: "All read/write (group and other write)",
		},
		{
			name:        "InvalidOwnerRWGroupOtherWrite",
			permissions: 0622,
			wantErr:     true,
			description: "Owner read/write, group/other write only",
		},
		{
			name:        "InvalidOwnerRWOtherWrite",
			permissions: 0602,
			wantErr:     true,
			description: "Owner read/write, other write",
		},
		{
			name:        "InvalidOwnerRWGroupWrite",
			permissions: 0620,
			wantErr:     true,
			description: "Owner read/write, group write",
		},
		{
			name:        "InvalidAllRWX",
			permissions: 0777,
			wantErr:     true,
			description: "All read/write/execute",
		},
		{
			name:        "InvalidGroupWriteOnly",
			permissions: 0020,
			wantErr:     true,
			description: "Group write only",
		},
		{
			name:        "InvalidOtherWriteOnly",
			permissions: 0002,
			wantErr:     true,
			description: "Other write only",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				BinDir:            "/tmp/bin",
				ScriptDir:         "/tmp/scripts",
				ScriptPermissions: tt.permissions,
			}

			err := cfg.Validate()

			if tt.wantErr && err == nil {
				t.Errorf("Validate() expected error for permissions 0%o (%s), but got nil", tt.permissions, tt.description)
			}

			if !tt.wantErr && err != nil {
				t.Errorf("Validate() unexpected error for permissions 0%o (%s): %v", tt.permissions, tt.description, err)
			}

			// Verify error message contains the permission value when error expected.
			if tt.wantErr && err != nil {
				expectedSubstring := "group and other write bits must not be set"
				if !contains(err.Error(), expectedSubstring) {
					t.Errorf("Validate() error message should contain %q, got: %v", expectedSubstring, err)
				}
			}
		})
	}
}

// TestConfigValidateDefault tests that default permissions (0600) are valid.
func TestConfigValidateDefault(t *testing.T) {
	cfg := &Config{
		BinDir:            "/tmp/bin",
		ScriptDir:         "/tmp/scripts",
		ScriptPermissions: 0600, // Default value
	}

	err := cfg.Validate()
	if err != nil {
		t.Errorf("Validate() should accept default permissions 0600, got error: %v", err)
	}
}

// contains checks if a string contains a substring.
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
