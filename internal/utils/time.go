package utils

import (
	"fmt"
	"time"
)

// Helper function to parse and format date strings to ISO format
func ParseISODate(dateStr string) (string, error) {
	t, err := time.Parse(time.RFC3339, dateStr)
	if err != nil {
		return "", fmt.Errorf("invalid date format: %w", err)
	}
	return t.Format(time.RFC3339), nil
}
