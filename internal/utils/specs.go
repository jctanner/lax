package utils

import (
	"strings"
)

type InstallSpec struct {
	Namespace    string
	Name         string
	Version      string
	Dependencies []InstallSpec
}

func (spec InstallSpec) Equals(other InstallSpec) bool {
	return spec.Namespace == other.Namespace &&
		spec.Name == other.Name &&
		spec.Version == other.Version
}

func SplitSpec(input string) []string {

	// geerlingguy.mac -> [geerlingguy, mac]
	// github.com:geerlingguy.mac -> []
	// https://github.com:geerlingguy.mac -> []
	// https://github.com/geerlingguy.mac -> []
	// git@github.com:geerlingguy/mac -> []

	colonIndex := strings.Index(input, ":")

	var result []string

	if colonIndex != -1 {
		// Check if the colon is part of a URL scheme
		if strings.HasPrefix(input, "http://") || strings.HasPrefix(input, "https://") {
			// Find the second colon if it exists
			secondColonIndex := strings.Index(input[colonIndex+1:], ":")
			if secondColonIndex != -1 {
				secondColonIndex += colonIndex + 1
				// Split at the second colon
				beforeColon := input[:secondColonIndex]
				afterColon := input[secondColonIndex+1:]

				// Split the part after the second colon on periods
				afterColonParts := strings.Split(afterColon, ".")

				// Combine the parts before the second colon and after the split
				result = append([]string{beforeColon}, afterColonParts...)
			} else {
				// If no second colon exists, treat the entire string as before the colon
				result = []string{input}
			}
		} else {
			// Split at the first colon
			beforeColon := input[:colonIndex]
			afterColon := input[colonIndex+1:]

			// Split the part after the first colon on periods
			afterColonParts := strings.Split(afterColon, ".")

			// Combine the parts before the colon and after the split
			result = append([]string{beforeColon}, afterColonParts...)
		}
	} else {
		// If no colon exists, split on all periods
		result = strings.Split(input, ".")
	}

	return result
}
