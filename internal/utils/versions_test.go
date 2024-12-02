package utils

import (
	"testing"

	"github.com/blang/semver/v4"
)

func TestFilterAndSortSemver(t *testing.T) {
	tests := []struct {
		name     string
		versions []string
		expected []string
	}{
		{
			name:     "Valid semantic versions",
			versions: []string{"1.0.0", "2.1.0", "1.2.3", "0.9.8"},
			expected: []string{"0.9.8", "1.0.0", "1.2.3", "2.1.0"},
		},
		{
			name:     "Mixed valid and invalid versions",
			versions: []string{"1.0.0", "invalid", "2.0.0", "not-a-version"},
			expected: []string{"1.0.0", "2.0.0"},
		},
		{
			name:     "No valid versions",
			versions: []string{"invalid", "not-a-version"},
			expected: []string{},
		},
		{
			name:     "Empty input",
			versions: []string{},
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, _ := FilterAndSortSemver(tt.versions)
			if !equalSlices(result, tt.expected) {
				t.Errorf("FilterAndSortSemver(%v) = %v, want %v", tt.versions, result, tt.expected)
			}
		})
	}
}

func TestGetHighestSemver(t *testing.T) {
	tests := []struct {
		name       string
		versions   []string
		expected   string
		expectErr  bool
	}{
		{
			name:      "Valid semantic versions",
			versions:  []string{"1.0.0", "2.1.0", "1.2.3", "0.9.8"},
			expected:  "2.1.0",
			expectErr: false,
		},
		{
			name:      "Mixed valid and invalid versions",
			versions:  []string{"1.0.0", "invalid", "2.0.0", "not-a-version"},
			expected:  "2.0.0",
			expectErr: false,
		},
		{
			name:      "No valid versions",
			versions:  []string{"invalid", "not-a-version"},
			expected:  "",
			expectErr: true,
		},
		{
			name:      "Empty input",
			versions:  []string{},
			expected:  "",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := GetHighestSemver(tt.versions)
			if (err != nil) != tt.expectErr {
				t.Errorf("GetHighestSemver(%v) error = %v, expectErr %v", tt.versions, err, tt.expectErr)
			}
			if result != tt.expected {
				t.Errorf("GetHighestSemver(%v) = %v, want %v", tt.versions, result, tt.expected)
			}
		})
	}
}

func TestCompareSemVersions(t *testing.T) {
	tests := []struct {
		name       string
		op         string
		v1         string
		v2         string
		expected   bool
		expectErr  bool
	}{
		{
			name:      "Greater than",
			op:        ">",
			v1:        "2.0.0",
			v2:        "1.0.0",
			expected:  true,
			expectErr: false,
		},
		{
			name:      "Less than or equal",
			op:        "<=",
			v1:        "1.0.0",
			v2:        "1.0.0",
			expected:  true,
			expectErr: false,
		},
		{
			name:      "Invalid operator",
			op:        "??",
			v1:        "1.0.0",
			v2:        "2.0.0",
			expected:  false,
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			version1, _ := semver.Parse(tt.v1)
			version2, _ := semver.Parse(tt.v2)
			result, err := CompareSemVersions(tt.op, &version1, &version2)
			if (err != nil) != tt.expectErr {
				t.Errorf("CompareSemVersions(%q, %q, %q) error = %v, expectErr %v", tt.op, tt.v1, tt.v2, err, tt.expectErr)
			}
			if result != tt.expected {
				t.Errorf("CompareSemVersions(%q, %q, %q) = %v, want %v", tt.op, tt.v1, tt.v2, result, tt.expected)
			}
		})
	}
}
