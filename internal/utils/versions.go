package utils

import (
	"fmt"
	"sort"

	"github.com/blang/semver/v4"
	//"github.com/blang/semver/v4"
	//"github.com/Masterminds/semver/v3"
)

func FilterAndSortSemver(versions []string) ([]string, error) {
	var validVersions []semver.Version

	// Parse and filter valid semantic versions
	for _, v := range versions {
		version, err := semver.Parse(v)
		if err == nil {
			validVersions = append(validVersions, version)
		}
	}

	// Sort the valid semantic versions
	sort.Slice(validVersions, func(i, j int) bool {
		return validVersions[i].LT(validVersions[j])
	})

	// Convert sorted versions back to strings
	var sortedVersionStrings []string
	for _, v := range validVersions {
		sortedVersionStrings = append(sortedVersionStrings, v.String())
	}

	return sortedVersionStrings, nil
}

func GetHighestSemver(versions []string) (string, error) {

	sorted, err := FilterAndSortSemver(versions)
	if err != nil {
		return "", err
	}

	if len(sorted) == 0 {
		return "", fmt.Errorf("no valid semantic versions")
	}

	return sorted[len(sorted)-1], nil

}

func CompareSemVersions(op string, v1 *semver.Version, v2 *semver.Version) (bool, error) {
	switch op {
	case ">":
		return v1.GT(*v2), nil
	case ">=":
		return v1.GT(*v2) || v1.EQ(*v2), nil
	case "<":
		return v1.LT(*v2), nil
	case "<=":
		return v1.LT(*v2) || v1.EQ(*v2), nil
	case "=":
		return v1.EQ(*v2), nil
	default:
		return false, fmt.Errorf("invalid operator: %s", op)
	}
}
