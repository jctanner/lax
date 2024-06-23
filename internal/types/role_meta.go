package types

import (
	"fmt"
	"strconv"
	"strings"
)

type RoleMeta struct {
	GalaxyInfo GalaxyInfo `yaml:"galaxy_info"`
}

type GalaxyInfo struct {
	Author    Author `yaml:"author"`
	Namespace string `yaml:"namespace"`
	RoleName  string `yaml:"role_name"`

	// this doesn't actually exist
	// but we want to have a settable
	// property for the index files
	Version string `yaml:"version"`

	Description       string           `yaml:"description"`
	License           RoleLicense      `yaml:"license"`
	MinAnsibleVersion string           `json:"min_ansible_version"`
	Platforms         []RolePlatform   `yaml:"platforms"`
	GalaxyTags        []GalaxyTags     `yaml:"galaxy_tags"`
	Dependencies      []RoleDependency `yaml:"dependencies"`
}

type Author struct {
	Value []string
}

func (a *Author) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var singleAuthor string
	if err := unmarshal(&singleAuthor); err == nil {
		a.Value = []string{singleAuthor}
		return nil
	}

	var multipleAuthors []string
	if err := unmarshal(&multipleAuthors); err == nil {
		a.Value = multipleAuthors
		return nil
	}

	return fmt.Errorf("failed to unmarshal Author field")
}

type RolePlatform struct {
	Name     string   `yaml:"name"`
	Versions []string `yaml:"versions"`
}

type RolePlatformVersions []string

func (rp *RolePlatform) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var temp struct {
		Name     string      `yaml:"name"`
		Versions interface{} `yaml:"versions"`
	}

	if err := unmarshal(&temp); err != nil {
		return err
	}

	rp.Name = temp.Name

	switch v := temp.Versions.(type) {
	case string:
		rp.Versions = RolePlatformVersions{v}
	case int:
		rp.Versions = RolePlatformVersions{strconv.Itoa(v)}
	case float64:
		vstring := strconv.FormatFloat(v, 'f', -1, 64)
		rp.Versions = RolePlatformVersions{vstring}
	case nil:
		rp.Versions = RolePlatformVersions{}
	case []interface{}:
		var versions RolePlatformVersions
		for _, version := range v {
			switch v := version.(type) {
			case string:
				versions = append(versions, v)
			case int:
				versions = append(versions, strconv.Itoa(v))
			case float64:
				vstring := strconv.FormatFloat(v, 'f', -1, 64)
				rp.Versions = append(versions, vstring)
			default:
				return fmt.Errorf("unexpected type for platform-version: %T", version)
			}
		}
		rp.Versions = versions
	default:
		return fmt.Errorf("unexpected type for platform-versions: %T", temp.Versions)
	}

	return nil
}

type RoleLicense []string

func (l *RoleLicense) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var singleLicense string
	if err := unmarshal(&singleLicense); err == nil {
		*l = RoleLicense{singleLicense}
		return nil
	}

	var licenseList []string
	if err := unmarshal(&licenseList); err == nil {
		*l = RoleLicense(licenseList)
		return nil
	}

	return fmt.Errorf("failed to unmarshal License")
}

// Dependency can be either a string or a map
type RoleDependency struct {
	Src  string
	Name string
}

func (d *RoleDependency) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var depString string
	if err := unmarshal(&depString); err == nil {
		d.Src = depString
		d.Name = depString
		return nil
	}

	var depMap map[string]string
	if err := unmarshal(&depMap); err == nil {
		d.Src = depMap["src"]
		d.Name = depMap["name"]
		return nil
	}

	return fmt.Errorf("failed to unmarshal Dependency")
}

type GalaxyTags []string

/*
func (gt *GalaxyTags) UnmarshalYAML(unmarshal func(interface{}) error) error {

	fmt.Printf("-----------------------------------")
	fmt.Printf("gt: %s\n", gt)
	fmt.Printf("-----------------------------------")
	var tags []string

	*gt = tags
	return nil

	// Try to unmarshal as a list of strings
	//var tags []string
	if err := unmarshal(&tags); err == nil {
		*gt = tags
		return nil
	}

	// If it fails, try to unmarshal as a single string and split it into a list
	var singleLine string
	if err := unmarshal(&singleLine); err == nil {
		*gt = strings.Fields(singleLine) // Split by whitespace
		return nil
	}

	// If both attempts fail, return an error
	return fmt.Errorf("failed to unmarshal GalaxyTags")
}
*/

/*
func (gt *GalaxyTags) UnmarshalYAML(unmarshal func(interface{}) error) error {

	fmt.Println("----------------------------------------------")
	fmt.Println("UnmarshalYAML called for GalaxyTags")
	fmt.Println("----------------------------------------------")

	// Try to unmarshal as a list of strings
	var tags []string
	if err := unmarshal(&tags); err == nil {
		*gt = tags
		return nil
	}

	// If it fails, try to unmarshal as a single string and split it into a list
	var singleLine string
	if err := unmarshal(&singleLine); err == nil {
		// Split by lines and trim spaces
		lines := strings.Split(singleLine, "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line != "" {
				*gt = append(*gt, line)
			}
		}
		return nil
	}

	// If both attempts fail, return an error
	return fmt.Errorf("failed to unmarshal GalaxyTags")
}
*/

// Implement custom unmarshaling for GalaxyTags
func (gt *GalaxyTags) UnmarshalYAML(unmarshal func(interface{}) error) error {

	// Try to unmarshal as a list of strings
	var tags []string
	if err := unmarshal(&tags); err == nil {
		*gt = tags
		return nil
	}

	// If it fails, try to unmarshal as a multiline string and split it into a list
	var singleLine string
	if err := unmarshal(&singleLine); err == nil {
		// Split by new lines and trim spaces
		lines := strings.Split(singleLine, "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line != "" {
				*gt = append(*gt, line)
			}
		}
		return nil
	}

	// If both attempts fail, return an error
	return fmt.Errorf("failed to unmarshal GalaxyTags")
}
