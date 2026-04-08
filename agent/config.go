package agent

import (
	"bytes"
	"os"

	"gopkg.in/yaml.v3"
)

type AgentConfig struct {
	ID string `yaml:"id"` //'auto' means detect hostname at runtime
}

type ServerConfig struct {
	Address string `yaml:"address"`
	TLS     bool   `yaml:"tls"`
}

type WatchConfig struct {
	Paths         []string `yaml:"paths"`
	Recursive     bool     `yaml:"recursive"`
	Exclude       []string `yaml:"exclude"`
	MaxFileSizeMB int      `yaml:"max_file_size_mb"`
}
type EventConfig struct {
	Include []string `yaml:"include"`
}
type ScanConfig struct {
	OnReconnect bool `yaml:"on_reconnect"`
}

// the root struct
type Config struct {
	Agent  AgentConfig  `yaml:"agent"`
	Server ServerConfig `yaml:"server"`
	Watch  WatchConfig  `yaml:"watch"`
	Events EventConfig  `yaml:"events"`
	Scan   ScanConfig   `yaml:"scan"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config Config

	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)

	if err := decoder.Decode(&config); err != nil {
		return nil, err
	}

	if config.Watch.MaxFileSizeMB == 0 {
		config.Watch.MaxFileSizeMB = 50 // Default to 50MB
	}

	return &config, nil
}
