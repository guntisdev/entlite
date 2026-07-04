package util

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type BufGenConfig struct {
	// ProtoTypesDir is the output directory for the generated Go proto types
	// (the buf.build/protocolbuffers/go plugin).
	ProtoTypesDir string
}

func GetBufConfigFromYaml(bufYamlPath string) (*BufGenConfig, error) {
	data, err := os.ReadFile(bufYamlPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read buf.gen.yaml: %w", err)
	}

	var config struct {
		Plugins []struct {
			Remote string `yaml:"remote"`
			Local  string `yaml:"local"`
			Out    string `yaml:"out"`
		} `yaml:"plugins"`
	}

	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse buf.gen.yaml: %w", err)
	}

	if len(config.Plugins) == 0 {
		return nil, fmt.Errorf("no plugins found in buf.gen.yaml")
	}

	for _, plugin := range config.Plugins {
		name := plugin.Remote
		if name == "" {
			name = plugin.Local
		}
		if strings.Contains(name, "protocolbuffers/go") {
			if plugin.Out == "" {
				return nil, fmt.Errorf("out not specified for protocolbuffers/go plugin in buf.gen.yaml")
			}
			return &BufGenConfig{
				ProtoTypesDir: plugin.Out,
			}, nil
		}
	}

	return nil, fmt.Errorf("no protocolbuffers/go plugin found in buf.gen.yaml")
}
