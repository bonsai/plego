package yaqqle

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Schema struct {
	Tables []Table `yaml:"tables" json:"tables"`
}

type Table struct {
	Name    string   `yaml:"name" json:"name"`
	Columns []Column `yaml:"columns" json:"columns"`
	Indexes []Index  `yaml:"indexes,omitempty" json:"indexes,omitempty"`
}

type Column struct {
	Name    string `yaml:"name" json:"name"`
	Type    string `yaml:"type" json:"type"`
	Nullable bool  `yaml:"nullable,omitempty" json:"nullable,omitempty"`
	PK      bool   `yaml:"pk,omitempty" json:"pk,omitempty"`
	Default string `yaml:"default,omitempty" json:"default,omitempty"`
	Unique  bool   `yaml:"unique,omitempty" json:"unique,omitempty"`
	JSONKey string `yaml:"json_key,omitempty" json:"json_key,omitempty"`
	Comment string `yaml:"comment,omitempty" json:"comment,omitempty"`
}

func (c Column) EffectiveJSONKey() string {
	if c.JSONKey != "" {
		return c.JSONKey
	}
	return c.Name
}

type Index struct {
	Name    string   `yaml:"name" json:"name"`
	Columns []string `yaml:"columns" json:"columns"`
	Unique  bool     `yaml:"unique,omitempty" json:"unique,omitempty"`
}

func ParseFile(path string) (*Schema, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var s Schema
	if err := yaml.Unmarshal(b, &s); err != nil {
		return nil, err
	}
	return &s, nil
}
