package util

import (
	"fmt"
	"os"

	"github.com/guntisdev/entlite/internal/generator/sqlc"
	"gopkg.in/yaml.v3"
)

type SqlcGenConfig struct {
	InputDir string
	Dialect  sqlc.SQLDialect
}

func GetSqlcConfigFromYaml(sqlcYamlPath string) (*SqlcGenConfig, error) {
	data, err := os.ReadFile(sqlcYamlPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read sqlc.yaml: %w", err)
	}

	var config struct {
		SQL []struct {
			Engine string `yaml:"engine"`
			Gen    struct {
				Go struct {
					Out string `yaml:"out"`
				} `yaml:"go"`
			} `yaml:"gen"`
		} `yaml:"sql"`
	}

	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse sqlc.yaml: %w", err)
	}

	if len(config.SQL) == 0 {
		return nil, fmt.Errorf("no SQL configurations found in sqlc.yaml")
	}

	engine := config.SQL[0].Engine
	if engine == "" {
		return nil, fmt.Errorf("engine not specified in sqlc.yaml")
	}

	dialect := sqlc.SQLDialect(engine)
	switch dialect {
	case sqlc.PostgreSQL, sqlc.MySQL, sqlc.SQLite:
	default:
		return nil, fmt.Errorf("unsupported SQL dialect '%s' in sqlc.yaml, must be one of: %s, %s, %s",
			engine, sqlc.PostgreSQL, sqlc.MySQL, sqlc.SQLite)
	}

	outputDir := config.SQL[0].Gen.Go.Out
	if outputDir == "" {
		return nil, fmt.Errorf("gen.go.out not specified in sqlc.yaml")
	}

	return &SqlcGenConfig{
		InputDir: outputDir,
		Dialect:  dialect,
	}, nil
}
