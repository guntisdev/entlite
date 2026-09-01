package docs

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDocsAreUpToDate(t *testing.T) {
	result, err := Run(Config{Root: repoRoot(t), OutDir: "docs", HTMLDir: "", Check: true})
	if err != nil {
		t.Fatalf("Failed to build docs: %v", err)
	}

	if len(result.Stale) > 0 {
		t.Fatalf("%s\n%d file(s) out of date, run: make docs", result.Diff, len(result.Stale))
	}
}

func repoRoot(t *testing.T) string {
	t.Helper()

	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to find working directory: %v", err)
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("Failed to find go.mod above the test directory")
		}
		dir = parent
	}
}
