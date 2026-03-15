package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/guntisdev/entlite/internal/util"
)

func TestProtoValidateCommand(t *testing.T) {

	tmpDir := t.TempDir()

	schemaDir := filepath.Join(tmpDir, "ent", "schema")
	logicDir := filepath.Join(tmpDir, "ent", "logic")
	pbDir := filepath.Join(tmpDir, "ent", "gen", "pb")

	if err := os.MkdirAll(schemaDir, 0755); err != nil {
		t.Fatalf("Failed to create schema directory: %v", err)
	}

	if err := os.MkdirAll(logicDir, 0755); err != nil {
		t.Fatalf("Failed to create logic directory: %v", err)
	}

	if err := os.MkdirAll(pbDir, 0755); err != nil {
		t.Fatalf("Failed to create pb directory: %v", err)
	}

	// Write common test input files
	writeTestGoMod(t, tmpDir)
	writeTestUserSchema(t, schemaDir)
	writeTestLogic(t, logicDir)

	// Change to tmpDir/ent to simulate working in the project
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	if err := os.Chdir(filepath.Join(tmpDir, "ent")); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}
	defer os.Chdir(originalDir)

	protoValidate([]string{pbDir})

	validatePath := filepath.Join(pbDir, "proto_validate.go")
	if _, err := os.Stat(validatePath); os.IsNotExist(err) {
		t.Fatalf("Expected proto_validate.go was not created at %s", validatePath)
	}

	actualContent, err := os.ReadFile(validatePath)
	if err != nil {
		t.Fatalf("Failed to read generated proto_validate.go: %v", err)
	}

	expectedContent := `package pb

import (
	"fmt"
	"github.com/guntisdev/entlite/examples/01-basic-entity/ent/logic"
)

func (r *CreateUserRequest) Validate() error {
	if !logic.StartsWithCapital(r.Name) {
		return fmt.Errorf("Validation failed for field name: Name")
	}
	return nil
}

func (r *UpdateUserRequest) Validate() error {
	if !logic.StartsWithCapital(r.Name) {
		return fmt.Errorf("Validation failed for field name: Name")
	}
	return nil
}`

	if d := util.Diff(expectedContent, string(actualContent)); d != "" {
		t.Errorf("proto_validate.go content mismatch (-expected +actual):\n%s", d)
	}
}
