package utils

import (
	"bufio"
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

func countLeadingWhitespaces(s string) int {
	re := regexp.MustCompile(`^\s*`)
	match := re.FindString(s)
	return len(match)
}

func AddLiteralBlockScalarToTags(yamlStr string) string {
	lines := strings.Split(yamlStr, "\n")
	modifiedLines := []string{}
	inGalaxyTags := false
	malformedTags := []string{}
	isList := false

	keyWhiteSpaceCount := 0
	leadingWhitespace := ""
	itemWhiteSpace := ""

	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)

		//fmt.Printf("%s\n", line)

		if strings.HasPrefix(trimmedLine, "galaxy_tags:") {
			keyWhiteSpaceCount = countLeadingSpaces(line)
			leadingWhitespace = strings.Repeat(" ", keyWhiteSpaceCount)
			itemWhiteSpace = strings.Repeat("  ", keyWhiteSpaceCount)
			inGalaxyTags = true
			malformedTags = append(malformedTags, strings.TrimSpace(strings.TrimPrefix(trimmedLine, "galaxy_tags:")))
			continue
		}

		// need to break if we find a new keyword ...
		//fmt.Printf("%b: %s\n", inGalaxyTags, line)
		if strings.Contains(line, "dependencies:") {
			inGalaxyTags = false
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
					//modifiedLines = append(modifiedLines, leadingWhitespace+"galaxy_tags: |")
					modifiedLines = append(modifiedLines, leadingWhitespace+"galaxy_tags:")
					for _, tag := range malformedTags {
						if tag != "" {
							cleanTag := strings.TrimSpace(tag)
							if strings.HasPrefix(cleanTag, "-") {
								modifiedLines = append(modifiedLines, itemWhiteSpace+cleanTag)
							} else {
								modifiedLines = append(modifiedLines, itemWhiteSpace+"- "+cleanTag)
							}
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
		//modifiedLines = append(modifiedLines, leadingWhitespace+"galaxy_tags: |")
		modifiedLines = append(modifiedLines, leadingWhitespace+"galaxy_tags:")
		for _, tag := range malformedTags {
			if tag != "" {
				//modifiedLines = append(modifiedLines, itemWhiteSpace+tag)
				cleanTag := strings.TrimSpace(tag)
				if strings.HasPrefix(cleanTag, "-") {
					modifiedLines = append(modifiedLines, itemWhiteSpace+cleanTag)
				} else {
					modifiedLines = append(modifiedLines, itemWhiteSpace+"- "+cleanTag)
				}
			}
		}
	}

	// If tags are in a list format, just return the original input
	if isList {
		return yamlStr
	}

	return strings.Join(modifiedLines, "\n") + "\n"
}

func FixPlatformVersion(yamlStr string) string {
	lines := strings.Split(yamlStr, "\n")
	modifiedLines := []string{}
	inPlatforms := false
	currentPlatformNameIndent := ""
	currentVersionIndent := ""

	for ix, line := range lines {
		trimmedLine := strings.TrimSpace(line)

		// uncomment platforms key?

		if strings.HasPrefix(trimmedLine, "platforms:") || strings.Contains(trimmedLine, "platforms:") {
			inPlatforms = true
			//if strings.HasPrefix(trimmedLine, "#") {
			//	line = strings.ReplaceAll(line, "#", "")
			//}
			//modifiedLines = append(modifiedLines, line)
			modifiedLines = append(modifiedLines, "  platforms:")
			continue
		}

		if inPlatforms {
			if strings.HasPrefix(trimmedLine, "- name:") {
				// Determine the indentation for the name and versions keys
				leadingSpaces := len(line) - len(strings.TrimLeft(line, " "))
				currentPlatformNameIndent = strings.Repeat(" ", leadingSpaces+2) // indent + 2 spaces for name level
				currentVersionIndent = strings.Repeat(" ", leadingSpaces+4)      // indent + 4 spaces for versions level
				modifiedLines = append(modifiedLines, line)

				// only works if they uncommented the line ...
				if !strings.Contains(lines[ix+1], "versions:") && !strings.Contains(lines[ix+2], "versions:") {
					modifiedLines = append(modifiedLines, currentPlatformNameIndent+"versions:")
				} else {
					// check for comments ...
					if strings.Contains(lines[ix+1], "versions:") && strings.Contains(lines[ix+1], "#") {
						nextLine := strings.TrimSpace(lines[ix+2])
						if nextLine != "" {
							modifiedLines = append(modifiedLines, currentPlatformNameIndent+"versions:")
						}
					} else if strings.Contains(lines[ix+2], "versions:") && strings.Contains(lines[ix+2], "#") {
						nextLine := strings.TrimSpace(lines[ix+3])
						if nextLine != "" {
							modifiedLines = append(modifiedLines, currentPlatformNameIndent+"versions:")
						}
					}
				}

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

	/*
		for ix, ml := range modifiedLines {
			fmt.Printf("ML(%d): %s\n", ix, ml)
		}
	*/

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

func ReplaceDependencyRoleWithName(yamlStr string) string {
	var result strings.Builder
	scanner := bufio.NewScanner(strings.NewReader(yamlStr))

	inDependencies := false

	for scanner.Scan() {
		line := scanner.Text()
		trimmedLine := strings.TrimSpace(line)

		if strings.HasPrefix(trimmedLine, "dependencies:") {
			inDependencies = true
		} else if inDependencies && (trimmedLine == "" || !strings.HasPrefix(trimmedLine, "-")) {
			inDependencies = false
		}

		if inDependencies && strings.Contains(trimmedLine, "role:") {
			line = strings.Replace(line, "role:", "name:", 1)
		}

		result.WriteString(line + "\n")
	}

	return result.String()
}

func RemoveDependenciesLiteralIfNoDeps(yamlStr string) string {

	// find the dependencies section?
	// does it have any non-empty traling lines?
	// does it have a trailing "|"?

	lines := strings.Split(yamlStr, "\n")

	hasLiteral := false
	hasInlineDeps := false
	hasInlineParens := false
	hasNextLineDeps := false
	dependencyKeyIndex := 0

	for ix, line := range lines {
		trimmedLine := strings.TrimSpace(line)

		if strings.HasPrefix(trimmedLine, "dependencies:") || strings.Contains(trimmedLine, "dependencies:") {
			dependencyKeyIndex = ix

			if strings.Contains(trimmedLine, "|") {
				hasLiteral = true
			}

			if trimmedLine != "depedencies:" {
				if strings.Contains(trimmedLine, "[]") {
					hasInlineParens = true
				} else {
					stripped := strings.Replace(trimmedLine, "|", "", 1)
					stripped = strings.TrimSpace(stripped)
					if stripped != "dependencies:" {
						hasInlineDeps = true
					}
				}
			} else if len(lines) > ix {
				if strings.TrimSpace(lines[ix+1]) != "" {
					hasNextLineDeps = true
				}
			}

			break
		}

	}

	fmt.Printf("FIXING %s\n", lines[dependencyKeyIndex])
	fmt.Printf("\thasLiteral: %b\n", hasLiteral)
	fmt.Printf("\thasInlineDeps: %s\n", hasInlineDeps)
	fmt.Printf("\thasNextLineDeps: %s\n", hasNextLineDeps)
	//fmt.Printf("\thasLiteral: %s\n", hasNextLineDeps)

	/*
		if !hasNextLineDeps && !hasLiteral && hasInlineDeps && !hasInlineParens {
			fixme := lines[dependencyKeyIndex]
			lines[dependencyKeyIndex] = strings.Replace(fixme, "|", "", 1)
		} else if hasInlineDeps && !hasNextLineDeps && !hasLiteral && !hasInlineParens {
			fixme := lines[dependencyKeyIndex]
			lines[dependencyKeyIndex] = strings.Replace(fixme, ":", ": | ", 1)
		} else if !hasInlineDeps && !hasInlineParens && !hasNextLineDeps && hasLiteral {
			fixme := lines[dependencyKeyIndex]
			lines[dependencyKeyIndex] = strings.Replace(fixme, "|", "", 1)
		}
	*/

	if !hasLiteral {
		if !hasNextLineDeps && hasInlineDeps && !hasInlineParens {
			fmt.Printf("##1 add newline + hyphen\n")
			fixme := lines[dependencyKeyIndex]
			lines[dependencyKeyIndex] = strings.Replace(fixme, ":", ":\n    -", 1)
		} else if hasInlineDeps && !hasNextLineDeps && !hasInlineParens {
			fmt.Printf("##2 add literal\n")
			fixme := lines[dependencyKeyIndex]
			lines[dependencyKeyIndex] = strings.Replace(fixme, ":", ": |\n    ", 1)
		}
	} else {
		if !hasInlineDeps && !hasNextLineDeps {
			fmt.Printf("##3 remove literal\n")
			fixme := lines[dependencyKeyIndex]
			lines[dependencyKeyIndex] = strings.Replace(fixme, "|", "", 1)
		}
	}

	return strings.Join(lines, "\n")
}

// AGGREGATE FIX IT ALL ...
func FixMetaMainYaml(yamlStr string) string {
	// trim all leadining and ending whitespace
	// find all keywords
	// indent keywords
	// for each keyword indent it's block
	// words after "dependencies:" should become a list
	return yamlStr
}
