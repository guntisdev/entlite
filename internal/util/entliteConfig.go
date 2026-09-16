package util

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type EntliteConfig struct {
	ProtoPackageName    string
	ProtoPackageVersion string
}

func GetEntliteConfigFromYaml(entliteYamlPath string) (*EntliteConfig, error) {
	data, err := os.ReadFile(entliteYamlPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read entlite.yaml: %w", err)
	}

	var config struct {
		Proto struct {
			Name    string `yaml:"name"`
			Version string `yaml:"version"`
		} `yaml:"proto"`
	}

	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse entlite.yaml: %w", err)
	}

	if config.Proto.Name == "" {
		return nil, fmt.Errorf("proto.name not specified in entlite.yaml")
	}
	if config.Proto.Version == "" {
		return nil, fmt.Errorf("proto.version not specified in entlite.yaml")
	}

	return &EntliteConfig{
		ProtoPackageName:    config.Proto.Name,
		ProtoPackageVersion: config.Proto.Version,
	}, nil
}
