package util

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type BufGenConfig struct {
	// ProtoTypesDir is where the buf.build/protocolbuffers/go plugin writes Go types
	ProtoTypesDir string
}

// localPlugin accepts buf's two forms for a "local" plugin
// `local: protoc-gen-go` vs `local: [go, run, some/pkg]`.
type localPlugin string

func (l *localPlugin) UnmarshalYAML(value *yaml.Node) error {
	switch value.Kind {
	case yaml.ScalarNode:
		var s string
		if err := value.Decode(&s); err != nil {
			return err
		}
		*l = localPlugin(s)
	case yaml.SequenceNode:
		var parts []string
		if err := value.Decode(&parts); err != nil {
			return err
		}
		*l = localPlugin(strings.Join(parts, " "))
	default:
		return fmt.Errorf("local: unsupported yaml node kind %v", value.Kind)
	}
	return nil
}

func GetBufConfigFromYaml(bufYamlPath string) (*BufGenConfig, error) {
	data, err := os.ReadFile(bufYamlPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read buf.gen.yaml: %w", err)
	}

	var config struct {
		Plugins []struct {
			Remote string      `yaml:"remote"`
			Local  localPlugin `yaml:"local"`
			Out    string      `yaml:"out"`
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
			name = string(plugin.Local)
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
