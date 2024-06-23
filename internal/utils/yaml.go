package utils

import (
	"fmt"
	"regexp"
	"strings"
)

// Helper function to count leading spaces
func countLeadingSpaces(line string) int {
	return len(line) - len(strings.TrimLeft(line, " "))
}

/*
Some meta/main.yml files have improperly independented "dependencies" keys.
*/
func FixGalaxyIndentation(yamlData string) (string, error) {
	lines := strings.Split(yamlData, "\n")
	galaxyInfoIndentation := -1
	indentationCount := make(map[int]int)

	// Identify the indentation level for the keys under galaxy_info
	for i, line := range lines {
		if strings.TrimSpace(line) == "galaxy_info:" {
			galaxyInfoIndentation = countLeadingSpaces(lines[i+1])
			continue
		}

		if galaxyInfoIndentation != -1 {
			if strings.HasPrefix(line, "  dependencies:") || strings.HasPrefix(line, "dependencies:") {
				//currentIndentation := countLeadingSpaces(line)
				lines[i] = strings.Repeat(" ", galaxyInfoIndentation) + strings.TrimSpace(line)
				continue
			}

			if strings.HasPrefix(line, "  ") && galaxyInfoIndentation != -1 {
				currentIndentation := countLeadingSpaces(line)
				indentationCount[currentIndentation]++
			}
		}
	}

	// Determine the most common indentation level
	mostCommonIndentation := galaxyInfoIndentation
	maxCount := 0
	for indent, count := range indentationCount {
		if count > maxCount {
			mostCommonIndentation = indent
			maxCount = count
		}
	}

	// Fix the indentation of the dependencies key to match the most common indentation
	for i, line := range lines {
		if strings.HasPrefix(line, "dependencies:") {
			lines[i] = strings.Repeat(" ", mostCommonIndentation) + "dependencies:"
		}
	}

	return strings.Join(lines, "\n"), nil
}

/*
Some descriptions might contain brackets which blows up the yaml parser
if not properly quoted. For example:

	[DRAFT] DISA STIG for Red Hat Virtualization Host (RHVH)
*/
func AddQuotesToDescription(yamlData string) string {
	lines := strings.Split(yamlData, "\n")
	for i, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		if strings.HasPrefix(trimmedLine, "description:") {
			// Extract the leading whitespace and the content of the description
			leadingWhitespace := line[:len(line)-len(trimmedLine)]
			description := strings.TrimSpace(strings.TrimPrefix(trimmedLine, "description:"))
			// Add quotes around the description content if it doesn't already have them
			if !(strings.HasPrefix(description, "\"") && strings.HasSuffix(description, "\"")) {
				lines[i] = fmt.Sprintf("%sdescription: \"%s\"", leadingWhitespace, description)
			}
		}
	}
	return strings.Join(lines, "\n")
}

/*
Sometimes a list is not a list ...
galaxy_tags:

	foo
	bar
	baz
*/
/*
func AddLiteralBlockScalarToTags(yamlStr string) string {
	lines := strings.Split(yamlStr, "\n")
	modifiedLines := []string{}
	inGalaxyTags := false
	malformedTags := []string{}

	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)

		if strings.HasPrefix(trimmedLine, "galaxy_tags:") {
			inGalaxyTags = true
			modifiedLines = append(modifiedLines, line)
			continue
		}

		if inGalaxyTags {
			match, _ := regexp.MatchString(`^\s*\w+`, trimmedLine)
			if match {
				malformedTags = append(malformedTags, strings.TrimSpace(line))
			} else {
				inGalaxyTags = false
				if len(malformedTags) > 0 {
					modifiedLines = append(modifiedLines, "galaxy_tags: |")
					for _, tag := range malformedTags {
						modifiedLines = append(modifiedLines, "  "+tag)
					}
					malformedTags = []string{}
				}
				modifiedLines = append(modifiedLines, line)
			}
		} else {
			modifiedLines = append(modifiedLines, line)
		}
	}

	// Handle case where galaxy_tags is at the end of the file
	if len(malformedTags) > 0 {
		modifiedLines = append(modifiedLines, "galaxy_tags: |")
		for _, tag := range malformedTags {
			modifiedLines = append(modifiedLines, "  "+tag)
		}
	}

	return strings.Join(modifiedLines, "\n")
}
*/

/*
func AddLiteralBlockScalarToTags(yamlStr string) string {
	lines := strings.Split(yamlStr, "\n")
	modifiedLines := []string{}
	inGalaxyTags := false
	malformedTags := []string{}

	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)

		if strings.HasPrefix(trimmedLine, "galaxy_tags:") {
			inGalaxyTags = true
			malformedTags = append(malformedTags, strings.TrimSpace(strings.TrimPrefix(trimmedLine, "galaxy_tags:")))
			continue
		}

		if inGalaxyTags {
			match, _ := regexp.MatchString(`^\s*\w+`, trimmedLine)
			if match {
				malformedTags = append(malformedTags, strings.TrimSpace(line))
			} else {
				inGalaxyTags = false
				if len(malformedTags) > 0 {
					modifiedLines = append(modifiedLines, "  galaxy_tags: |")
					for _, tag := range malformedTags {
						if tag != "" {
							modifiedLines = append(modifiedLines, "  "+tag)
						}
					}
					malformedTags = []string{}
				}
				modifiedLines = append(modifiedLines, line)
			}
		} else {
			modifiedLines = append(modifiedLines, line)
		}
	}

	// Handle case where galaxy_tags is at the end of the file
	if len(malformedTags) > 0 {
		modifiedLines = append(modifiedLines, "  galaxy_tags: |")
		for _, tag := range malformedTags {
			if tag != "" {
				modifiedLines = append(modifiedLines, "  "+tag)
			}
		}
	}

	return strings.Join(modifiedLines, "\n")
}
*/

func AddLiteralBlockScalarToTags(yamlStr string) string {
	lines := strings.Split(yamlStr, "\n")
	modifiedLines := []string{}
	inGalaxyTags := false
	malformedTags := []string{}
	isList := false

	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)

		if strings.HasPrefix(trimmedLine, "galaxy_tags:") {
			inGalaxyTags = true
			malformedTags = append(malformedTags, strings.TrimSpace(strings.TrimPrefix(trimmedLine, "galaxy_tags:")))
			continue
		}

		if inGalaxyTags {
			if strings.HasPrefix(trimmedLine, "-") {
				// If any line starts with a hyphen, assume it's a proper list and skip processing
				isList = true
				break
			}
			match, _ := regexp.MatchString(`^\s*\w+`, trimmedLine)
			if match {
				malformedTags = append(malformedTags, strings.TrimSpace(line))
			} else {
				inGalaxyTags = false
				if len(malformedTags) > 0 {
					modifiedLines = append(modifiedLines, "galaxy_tags: |")
					for _, tag := range malformedTags {
						if tag != "" {
							modifiedLines = append(modifiedLines, "  "+tag)
						}
					}
					malformedTags = []string{}
				}
				modifiedLines = append(modifiedLines, line)
			}
		} else {
			modifiedLines = append(modifiedLines, line)
		}
	}

	// Handle case where galaxy_tags is at the end of the file
	if len(malformedTags) > 0 && !isList {
		modifiedLines = append(modifiedLines, "galaxy_tags: |")
		for _, tag := range malformedTags {
			if tag != "" {
				modifiedLines = append(modifiedLines, "  "+tag)
			}
		}
	}

	// If tags are in a list format, just return the original input
	if isList {
		return yamlStr
	}

	return strings.Join(modifiedLines, "\n")
}

func FixPlatformVersion(yamlStr string) string {
	lines := strings.Split(yamlStr, "\n")
	modifiedLines := []string{}
	inPlatforms := false
	currentPlatformNameIndent := ""
	currentVersionIndent := ""

	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)

		if strings.HasPrefix(trimmedLine, "platforms:") {
			inPlatforms = true
			modifiedLines = append(modifiedLines, line)
			continue
		}

		if inPlatforms {
			if strings.HasPrefix(trimmedLine, "- name:") {
				// Determine the indentation for the name and versions keys
				leadingSpaces := len(line) - len(strings.TrimLeft(line, " "))
				currentPlatformNameIndent = strings.Repeat(" ", leadingSpaces+2) // indent + 2 spaces for name level
				currentVersionIndent = strings.Repeat(" ", leadingSpaces+4)      // indent + 4 spaces for versions level
				modifiedLines = append(modifiedLines, line)
				continue
			}

			if strings.HasPrefix(trimmedLine, "versions:") {
				modifiedLines = append(modifiedLines, currentPlatformNameIndent+"versions:")
				continue
			}

			// Add the version items with proper indentation
			if strings.HasPrefix(trimmedLine, "- all") && currentVersionIndent != "" {
				modifiedLines = append(modifiedLines, currentVersionIndent+"- all")
				continue
			}

			// Check for the end of the platforms section
			if trimmedLine == "" || strings.HasPrefix(trimmedLine, "categories:") {
				inPlatforms = false
				currentPlatformNameIndent = ""
				currentVersionIndent = ""
			}
		}

		modifiedLines = append(modifiedLines, line)
	}

	return strings.Join(modifiedLines, "\n")
}

func RemoveComments(yamlStr string) string {
	lines := strings.Split(yamlStr, "\n")
	modifiedLines := []string{}

	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		if strings.HasPrefix(trimmedLine, "#") {
			continue
		}
		modifiedLines = append(modifiedLines, line)
	}

	return strings.Join(modifiedLines, "\n")
}
