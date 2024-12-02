//go:build unit

package utils

import (
	"testing"
)

func TestParseISODate(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    string
		expectError bool
	}{
		{
			name:        "Valid ISO Date",
			input:       "2023-11-30T15:04:05Z",
			expected:    "2023-11-30T15:04:05Z",
			expectError: false,
		},
		//{
		//	name:        "Valid ISO Date with Timezone",
		//	input:       "2023-11-30T15:04:05+01:00",
		//	expected:    "2023-11-30T14:04:05Z", // Adjusted to UTC
		//	expectError: false,
		//},
		{
			name:        "Invalid Date Format",
			input:       "30-11-2023",
			expected:    "",
			expectError: true,
		},
		{
			name:        "Empty Date String",
			input:       "",
			expected:    "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseISODate(tt.input)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected an error, got none")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if result != tt.expected {
					t.Errorf("expected %q, got %q", tt.expected, result)
				}
			}
		})
	}
}
