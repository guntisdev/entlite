package util

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/guntisdev/entlite/internal/generator/sqlc"
)

func TestGetSqlcConfigFromYaml_Success(t *testing.T) {
	tests := []struct {
		name             string
		yamlContent      string
		expectedDialect  sqlc.SQLDialect
		expectedInputDir string
	}{
		{
			name: "postgresql dialect",
			yamlContent: `version: "2"
sql:
  - schema: "contract/sqlc/schema.sql"
    queries: "contract/sqlc/queries.sql"    
    engine: "postgresql"
    gen:
      go:
        package: "internal"
        out: "gen/db/internal"
        emit_json_tags: true`,
			expectedDialect:  sqlc.PostgreSQL,
			expectedInputDir: "gen/db/internal",
		},
		{
			name: "sqlite dialect",
			yamlContent: `version: "2"
sql:
  - schema: "contract/sqlc/schema.sql"
    queries: "contract/sqlc/queries.sql"    
    engine: "sqlite"
    gen:
      go:
        package: "internal"
        out: "gen/db/internal"
        emit_json_tags: true`,
			expectedDialect:  sqlc.SQLite,
			expectedInputDir: "gen/db/internal",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			tmpFile := filepath.Join(tmpDir, "sqlc.yaml")

			if err := os.WriteFile(tmpFile, []byte(tt.yamlContent), 0644); err != nil {
				t.Fatalf("Failed to create temp file: %v", err)
			}

			config, err := GetSqlcConfigFromYaml(tmpFile)
			if err != nil {
				t.Fatalf("Expected no error, got: %v", err)
			}

			if config.Dialect != tt.expectedDialect {
				t.Errorf("Expected dialect %s, got %s", tt.expectedDialect, config.Dialect)
			}

			if config.InputDir != tt.expectedInputDir {
				t.Errorf("Expected input dir %s, got %s", tt.expectedInputDir, config.InputDir)
			}
		})
	}
}

func TestGetSqlcConfigFromYaml_Failures(t *testing.T) {
	tests := []struct {
		name        string
		yamlContent string
		errorMsg    string
	}{
		{
			name: "invalid dialect",
			yamlContent: `version: "2"
sql:
  - schema: "contract/sqlc/schema.sql"
    queries: "contract/sqlc/queries.sql"    
    engine: "randomString"
    gen:
      go:
        package: "internal"
        out: "gen/db/internal"
        emit_json_tags: true`,
			errorMsg: "unsupported SQL dialect",
		},
		{
			name: "missing engine field",
			yamlContent: `version: "2"
sql:
  - schema: "contract/sqlc/schema.sql"
    queries: "contract/sqlc/queries.sql"
    gen:
      go:
        package: "internal"
        out: "gen/db/internal"
        emit_json_tags: true`,
			errorMsg: "engine not specified",
		},
		{
			name: "empty sql array",
			yamlContent: `version: "2"
sql: []`,
			errorMsg: "no SQL configurations found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			tmpFile := filepath.Join(tmpDir, "sqlc.yaml")

			if err := os.WriteFile(tmpFile, []byte(tt.yamlContent), 0644); err != nil {
				t.Fatalf("Failed to create temp file: %v", err)
			}

			_, err := GetSqlcConfigFromYaml(tmpFile)
			if err == nil {
				t.Fatal("Expected error, got none")
			}
		})
	}
}
