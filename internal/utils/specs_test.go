package utils

import (
	"reflect"
	"testing"
)

func TestInstallSpecEquals(t *testing.T) {
	tests := []struct {
		name     string
		spec1    InstallSpec
		spec2    InstallSpec
		expected bool
	}{
		{
			name: "Equal Specs",
			spec1: InstallSpec{
				Namespace: "namespace",
				Name:      "name",
				Version:   "1.0.0",
			},
			spec2: InstallSpec{
				Namespace: "namespace",
				Name:      "name",
				Version:   "1.0.0",
			},
			expected: true,
		},
		{
			name: "Different Namespace",
			spec1: InstallSpec{
				Namespace: "namespace1",
				Name:      "name",
				Version:   "1.0.0",
			},
			spec2: InstallSpec{
				Namespace: "namespace2",
				Name:      "name",
				Version:   "1.0.0",
			},
			expected: false,
		},
		{
			name: "Different Name",
			spec1: InstallSpec{
				Namespace: "namespace",
				Name:      "name1",
				Version:   "1.0.0",
			},
			spec2: InstallSpec{
				Namespace: "namespace",
				Name:      "name2",
				Version:   "1.0.0",
			},
			expected: false,
		},
		{
			name: "Different Version",
			spec1: InstallSpec{
				Namespace: "namespace",
				Name:      "name",
				Version:   "1.0.0",
			},
			spec2: InstallSpec{
				Namespace: "namespace",
				Name:      "name",
				Version:   "2.0.0",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.spec1.Equals(tt.spec2)
			if result != tt.expected {
				t.Errorf("InstallSpec.Equals() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestSplitSpec(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "Simple string without colon",
			input:    "geerlingguy.mac",
			expected: []string{"geerlingguy", "mac"},
		},
		{
			name:     "Colon with period-separated string",
			input:    "github.com:geerlingguy.mac",
			expected: []string{"github.com", "geerlingguy", "mac"},
		},
        /*
		{
			name:     "URL with colon-separated string",
			input:    "https://github.com:geerlingguy.mac",
			expected: []string{"https://github.com:geerlingguy", "mac"},
		},
		{
			name:     "URL without colon-separated string",
			input:    "https://github.com/geerlingguy.mac",
			expected: []string{"https://github.com/geerlingguy", "mac"},
		},
		{
			name:     "Git SSH URL",
			input:    "git@github.com:geerlingguy/mac",
			expected: []string{"git@github.com", "geerlingguy", "mac"},
		},
        */
		{
			name:     "String without periods",
			input:    "github",
			expected: []string{"github"},
		},
		{
			name:     "Empty string",
			input:    "",
			expected: []string{""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SplitSpec(tt.input)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("SplitSpec(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}
