package types

import (
	"fmt"
	"strconv"
	"strings"
)

type rawYAML struct {
	Data interface{}
}

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
	Src     string
	Name    string
	Version string
}

func (d *RoleDependency) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var raw interface{}
	if err := unmarshal(&raw); err != nil {
		return err
	}

	switch data := raw.(type) {
	case string:
		//data = strings.Replace(data, "role:", "name:", 1)
		d.Src = data
		d.Name = data
		d.Version = ""
	case map[interface{}]interface{}:
		if src, ok := data["src"].(string); ok {
			d.Src = src
		}
		if name, ok := data["name"].(string); ok {
			d.Name = name
		}
		if version, ok := data["version"].(string); ok {
			d.Version = version
		}
	default:
		return fmt.Errorf("unexpected type: %T", data)
	}
	return nil
}

type GalaxyTags []string

// Implement custom unmarshaling for GalaxyTags
func (gt *GalaxyTags) UnmarshalYAML(unmarshal func(interface{}) error) error {

	//fmt.Println("************************************")

	var raw rawYAML
	if err := unmarshal(&raw); err != nil {
		fmt.Printf("ERROR %s\n", err)
		return err
	}

	// Try to unmarshal as a list of strings
	var tags []string
	if err := unmarshal(&tags); err == nil {
		//fmt.Printf("A: %s\n", tags)
		*gt = tags
		return nil
	}

	// If it fails, try to unmarshal as a multiline string and split it into a list
	var singleLine string
	if err := unmarshal(&singleLine); err == nil {
		// Split by new lines and trim spaces
		lines := strings.Split(singleLine, "\n")
		//fmt.Printf("B: %s\n", lines)
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

func (r *rawYAML) UnmarshalYAML(unmarshal func(interface{}) error) error {
	//fmt.Printf("***********************************\n")
	var raw interface{}
	if err := unmarshal(&raw); err != nil {
		//fmt.Print("ERROR DOING RAW UNMARSHALL\n")
		return err
	}
	//fmt.Printf("RAW: %s\n", raw)
	r.Data = raw
	return nil
}
