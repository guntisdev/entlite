package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewCommandFunction(t *testing.T) {
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(tmpDir)

	newCommand([]string{"User", "Post"})

	expectedFiles := []string{
		"ent/schema/user.go",
		"ent/schema/post.go",
		"ent/schema/generate.go",
		"ent/sqlc.yaml",
		"ent/buf.yaml",
		"ent/buf.gen.yaml",
		"ent/generate.go",
	}

	// Check all expected files exist
	for _, file := range expectedFiles {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			t.Errorf("Expected file %s was not created", file)
		}
	}

	// Check no unexpected files were created
	var actualFiles []string
	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			cleanPath := strings.TrimPrefix(filepath.ToSlash(path), "./")
			actualFiles = append(actualFiles, cleanPath)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("Failed to walk directory: %v", err)
	}

	expectedMap := make(map[string]bool)
	for _, file := range expectedFiles {
		expectedMap[file] = true
	}

	for _, actual := range actualFiles {
		if !expectedMap[actual] {
			t.Errorf("Unexpected file created: %s", actual)
		}
	}
}

func TestGenCommandFunction(t *testing.T) {
	t.Skip("Skipping genCommand test until implementation is complete")
}
