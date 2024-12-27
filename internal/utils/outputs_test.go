package utils

import (
	"testing"
)

func TestPrettyPrint(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected string
		wantErr  bool
	}{
		{
			name: "Simple map",
			input: map[string]string{
				"key1": "value1",
				"key2": "value2",
			},
			expected: `{
  "key1": "value1",
  "key2": "value2"
}`,
			wantErr: false,
		},
		{
			name:     "Empty map",
			input:    map[string]string{},
			expected: `{}`,
			wantErr:  false,
		},
		{
			name:     "Nil input",
			input:    nil,
			expected: `null`,
			wantErr:  false,
		},
		{
			name: "Complex struct",
			input: struct {
				Name  string
				Age   int
				Items []string
			}{
				Name:  "Alice",
				Age:   30,
				Items: []string{"item1", "item2"},
			},
			expected: `{
  "Name": "Alice",
  "Age": 30,
  "Items": [
    "item1",
    "item2"
  ]
}`,
			wantErr: false,
		},
		{
			name:     "Invalid input (unsupported type)",
			input:    make(chan int),
			expected: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := PrettyPrint(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("PrettyPrint() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if result != tt.expected && !tt.wantErr {
				t.Errorf("PrettyPrint() = %q, want %q", result, tt.expected)
			}
		})
	}
}
