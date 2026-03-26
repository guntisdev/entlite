package util

import (
	"fmt"
	"os"

	"github.com/guntisdev/entlite/internal/generator/sqlc"
	"gopkg.in/yaml.v3"
)

type SqlcConfig struct {
	Version string      `yaml:"version"`
	SQL     []SqlConfig `yaml:"sql"`
}

type SqlConfig struct {
	Schema  string    `yaml:"schema"`
	Queries string    `yaml:"queries"`
	Engine  string    `yaml:"engine"`
	Gen     GenConfig `yaml:"gen"`
}

type GenConfig struct {
	Go GoConfig `yaml:"go"`
}

type GoConfig struct {
	Package      string `yaml:"package"`
	Out          string `yaml:"out"`
	EmitJsonTags bool   `yaml:"emit_json_tags"`
}

func GetSqlDialectFromSqlcYaml(sqlcYamlPath string) (sqlc.SQLDialect, error) {
	data, err := os.ReadFile(sqlcYamlPath)
	if err != nil {
		return "", fmt.Errorf("failed to read sqlc.yaml: %w", err)
	}

	var config SqlcConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return "", fmt.Errorf("failed to parse sqlc.yaml: %w", err)
	}

	if len(config.SQL) == 0 {
		return "", fmt.Errorf("no SQL configurations found in sqlc.yaml")
	}

	engine := config.SQL[0].Engine
	if engine == "" {
		return "", fmt.Errorf("engine not specified in sqlc.yaml")
	}

	dialect := sqlc.SQLDialect(engine)
	switch dialect {
	case sqlc.PostgreSQL, sqlc.MySQL, sqlc.SQLite:
		return dialect, nil
	default:
		return "", fmt.Errorf("unsupported SQL dialect '%s' in sqlc.yaml, must be one of: %s, %s, %s",
			engine, sqlc.PostgreSQL, sqlc.MySQL, sqlc.SQLite)
	}
}

func GetSqlcOutputDirFromYaml(sqlcYamlPath string) (string, error) {
	data, err := os.ReadFile(sqlcYamlPath)
	if err != nil {
		return "", fmt.Errorf("failed to read sqlc.yaml: %w", err)
	}

	var config SqlcConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return "", fmt.Errorf("failed to parse sqlc.yaml: %w", err)
	}

	if len(config.SQL) == 0 {
		return "", fmt.Errorf("no SQL configurations found in sqlc.yaml")
	}

	outputDir := config.SQL[0].Gen.Go.Out
	if outputDir == "" {
		return "", fmt.Errorf("gen.go.out not specified in sqlc.yaml")
	}

	return outputDir, nil
}
