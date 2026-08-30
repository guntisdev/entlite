package examples

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const (
	teachesStart = "<!-- teaches:start -->"
	teachesEnd   = "<!-- teaches:end -->"
)

// fails when an example has no README.md or an empty teaches block
func TestExampleReadmes(t *testing.T) {
	dirs, err := exampleDirs(".")
	if err != nil {
		t.Fatalf("Failed to find examples: %v", err)
	}
	if len(dirs) == 0 {
		t.Fatal("No examples found")
	}

	for _, dir := range dirs {
		t.Run(dir, func(t *testing.T) {
			readme := filepath.Join(dir, "README.md")
			content, err := os.ReadFile(readme)
			if err != nil {
				t.Fatalf("Missing %s, every example needs one: %v", readme, err)
			}

			bullets, err := teaches(string(content))
			if err != nil {
				t.Fatalf("%s: %v", readme, err)
			}
			if len(bullets) == 0 {
				t.Fatalf("%s has an empty teaches block, list what the example teaches", readme)
			}
		})
	}
}

// returns the bullet lines between the two markers.
func teaches(content string) ([]string, error) {
	_, after, found := strings.Cut(content, teachesStart)
	if !found {
		return nil, missing(teachesStart)
	}
	block, _, found := strings.Cut(after, teachesEnd)
	if !found {
		return nil, missing(teachesEnd)
	}

	var bullets []string
	for line := range strings.SplitSeq(block, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "- ") {
			bullets = append(bullets, line)
		}
	}

	return bullets, nil
}

func missing(marker string) error {
	return &os.PathError{Op: "find", Path: marker, Err: os.ErrNotExist}
}

func exampleDirs(root string) ([]string, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, err
	}

	var found []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		dir := filepath.Join(root, entry.Name())
		if hasDialect(dir) {
			found = append(found, entry.Name())
		}
	}

	return found, nil
}

func hasDialect(dir string) bool {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		if _, err := os.Stat(filepath.Join(dir, entry.Name(), "ent", "generate.go")); err == nil {
			return true
		}
	}

	return false
}
