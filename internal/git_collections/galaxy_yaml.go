package git_collections

import (
	"fmt"
	"io/ioutil"

	"os"

	"gopkg.in/yaml.v2"
)

// Struct to represent the expected structure of the YAML file
type Config struct {
	Dependencies map[string]string `yaml:"dependencies"`
}

// readDependencies reads a YAML file and extracts the "dependencies" key
func ReadDependencies(filePath string) (map[string]string, error) {
	// Read the file
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %v", err)
	}

	// Parse the YAML content
	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %v", err)
	}

	return config.Dependencies, nil
}

// parseYAMLFile parses a YAML file into a map with a key "collections" that contains a list of maps.
func ParseRequirementsYAMLFile(filePath string) (map[string]interface{}, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %v", err)
	}

	var result map[string]interface{}
	err = yaml.Unmarshal(data, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal YAML: %v", err)
	}

	// Validate the structure
	collections, ok := result["collections"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("key 'collections' is missing or not a list")
	}

	for _, collection := range collections {
		collectionMap, ok := collection.(map[interface{}]interface{})
		if !ok {
			return nil, fmt.Errorf("collections should be a list of maps")
		}

		for _, key := range []string{"name", "source", "version"} {
			if _, ok := collectionMap[key]; !ok {
				return nil, fmt.Errorf("each collection should contain the key '%s'", key)
			}
		}
	}

	return result, nil
}
