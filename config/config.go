package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Pipeline PipelineConfig `yaml:"pipeline"`
}

type PipelineConfig struct {
	Source  SourceConfig   `yaml:"source"`
	Outputs []OutputConfig `yaml:"outputs"`
	StateDB string         `yaml:"state_db"`
}

type SourceConfig struct {
	Module     string   `yaml:"module"`
	Path       string   `yaml:"path"`
	Extensions []string `yaml:"extensions"`
}

type OutputConfig struct {
	Module      string `yaml:"module"`
	To          string `yaml:"to"`
	Credentials string `yaml:"credentials"`
	Token       string `yaml:"token"`
}

func Load(path string) (*Config, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	return &cfg, yaml.Unmarshal(b, &cfg)
}
