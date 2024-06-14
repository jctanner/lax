package galaxy_sync

import (
	"os"

	"gopkg.in/yaml.v2"
)

// Requirements represents the structure of requirements.yml
type Requirements struct {
	Collections []CollectionRequirement `yaml:"collections"`
	Roles       []RoleRequirement       `yaml:"roles"`
}

// Collection represents an Ansible collection in requirements.yml
type CollectionRequirement struct {
	Name    string `yaml:"name"`
	Version string `yaml:"version,omitempty"`
}

// Role represents an Ansible role in requirements.yml
type RoleRequirement struct {
	Name    string `yaml:"name"`
	Src     string `yaml:"src,omitempty"`
	Version string `yaml:"version,omitempty"`
}

func parseRequirements(filename string) (*Requirements, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var requirements Requirements
	err = yaml.Unmarshal(data, &requirements)
	if err != nil {
		return nil, err
	}

	return &requirements, nil
}
