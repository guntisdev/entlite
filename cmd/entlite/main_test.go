package main

import (
	"os"
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
	}

	for _, file := range expectedFiles {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			t.Errorf("Expected file %s was not created", file)
		}
	}
}

func TestGenCommandFunction(t *testing.T) {
	t.Skip("Skipping genCommand test until implementation is complete")
}
