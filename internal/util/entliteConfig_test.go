package util

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestGetEntliteConfigFromYaml_Success(t *testing.T) {
	yamlContent := `proto:
  name: entlite
  version: v1
`
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "entlite.yaml")

	if err := os.WriteFile(tmpFile, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	config, err := GetEntliteConfigFromYaml(tmpFile)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if config.ProtoPackageName != "entlite" {
		t.Errorf("Expected proto package name 'entlite', got %s", config.ProtoPackageName)
	}
	if config.ProtoPackageVersion != "v1" {
		t.Errorf("Expected proto package version 'v1', got %s", config.ProtoPackageVersion)
	}
}

func TestGetEntliteConfigFromYaml_Failures(t *testing.T) {
	tests := []struct {
		name        string
		yamlContent string
	}{
		{
			name: "missing name field",
			yamlContent: `proto:
  version: v1
`,
		},
		{
			name: "missing version field",
			yamlContent: `proto:
  name: entlite
`,
		},
		{
			name:        "empty file",
			yamlContent: ``,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			tmpFile := filepath.Join(tmpDir, "entlite.yaml")

			if err := os.WriteFile(tmpFile, []byte(tt.yamlContent), 0644); err != nil {
				t.Fatalf("Failed to create temp file: %v", err)
			}

			_, err := GetEntliteConfigFromYaml(tmpFile)
			if err == nil {
				t.Fatal("Expected error, got none")
			}
		})
	}
}

func TestGetEntliteConfigFromYaml_FileNotFound(t *testing.T) {
	_, err := GetEntliteConfigFromYaml(filepath.Join(t.TempDir(), "entlite.yaml"))
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("Expected a not-exist error, got: %v", err)
	}
}
