package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Global  GlobalConfig  `yaml:"global"`
	Plugins []PluginConfig `yaml:"plugins"`
}

type GlobalConfig struct {
	Timezone string `yaml:"timezone"`
	StateDB  string `yaml:"state_db"`
}

type PluginConfig struct {
	Module string                 `yaml:"module"`
	Config map[string]interface{} `yaml:"config,omitempty"`
}

func Load(path string) (*Config, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	return &cfg, yaml.Unmarshal(b, &cfg)
}
