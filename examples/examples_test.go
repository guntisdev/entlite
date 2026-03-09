//go:build integration
// +build integration

package examples

import (
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/guntisdev/entlite/internal/util"
)

var entliteBinary string

func TestMain(m *testing.M) {
	requiredTools := []string{"sqlc", "buf"}
	for _, tool := range requiredTools {
		if _, err := exec.LookPath(tool); err != nil {
			panic("Required tool '" + tool + "' is not installed. " +
				"Integration tests require sqlc and buf to be installed.\n" +
				"Install sqlc: https://docs.sqlc.dev/en/latest/overview/install.html\n" +
				"Install buf: https://buf.build/docs/installation\n")
		}
	}

	// Build entlite binary before running tests
	tmpDir, err := os.MkdirTemp("", "entlite-test-")
	if err != nil {
		panic("Failed to create temp dir: " + err.Error())
	}
	defer os.RemoveAll(tmpDir)

	entliteBinary = filepath.Join(tmpDir, "entlite")

	// Get module root (parent of examples/)
	wd, err := os.Getwd()
	if err != nil {
		panic("Failed to get working directory: " + err.Error())
	}
	moduleRoot := filepath.Dir(wd)

	cmd := exec.Command("go", "build", "-o", entliteBinary, "./cmd/entlite")
	cmd.Dir = moduleRoot
	if output, err := cmd.CombinedOutput(); err != nil {
		panic("Failed to build entlite: " + err.Error() + "\n" + string(output))
	}

	// Run tests
	code := m.Run()
	os.Exit(code)
}

func TestExamples(t *testing.T) {
	// Auto-discover all example directories
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("Failed to read examples directory: %v", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		schemaDir := filepath.Join(entry.Name(), "ent", "schema")
		if _, err := os.Stat(schemaDir); err != nil {
			continue
		}

		t.Run(entry.Name(), func(t *testing.T) {
			testExample(t, entry.Name())
		})
	}
}

func testExample(t *testing.T, exampleDir string) {
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	moduleRoot := filepath.Dir(originalDir) // examples/.. = module root

	tmpDir := t.TempDir()

	// Copy the entire example to temp directory
	srcDir := filepath.Join(".", exampleDir)
	dstDir := filepath.Join(tmpDir, exampleDir)

	if err := copyDir(srcDir, dstDir); err != nil {
		t.Fatalf("Failed to copy example to temp dir: %v", err)
	}

	// Create a go.mod file with replace directive pointing to local entlite
	goModContent := "module github.com/guntisdev/entlite/examples/" + exampleDir + "\n\n" +
		"go 1.25.6\n\n" +
		"replace github.com/guntisdev/entlite => " + moduleRoot + "\n"
	goModPath := filepath.Join(dstDir, "go.mod")
	if err := os.WriteFile(goModPath, []byte(goModContent), 0644); err != nil {
		t.Fatalf("Failed to create go.mod: %v", err)
	}

	// Update generate.go files to use the built binary instead of go run
	generateFiles := []string{
		filepath.Join(dstDir, "ent", "generate.go"),
		filepath.Join(dstDir, "ent", "schema", "generate.go"),
	}

	sqlcPath, _ := exec.LookPath("sqlc")
	bufPath, _ := exec.LookPath("buf")

	for _, generatePath := range generateFiles {
		if _, err := os.Stat(generatePath); os.IsNotExist(err) {
			continue // Skip if file doesn't exist
		}

		generateContent, err := os.ReadFile(generatePath)
		if err != nil {
			t.Fatalf("Failed to read %s: %v", generatePath, err)
		}

		// Replace patterns:
		// 1. "go run github.com/guntisdev/entlite/cmd/entlite" -> entlite binary
		// 2. "go run ../../../../cmd/entlite" -> entlite binary
		// 3. "go tool sqlc" -> sqlc binary
		// 4. "go tool buf" -> buf binary
		updatedContent := strings.ReplaceAll(string(generateContent),
			"go run github.com/guntisdev/entlite/cmd/entlite",
			entliteBinary)
		updatedContent = strings.ReplaceAll(updatedContent,
			"go run ../../../../cmd/entlite",
			entliteBinary)
		updatedContent = strings.ReplaceAll(updatedContent,
			"go tool sqlc",
			sqlcPath)
		updatedContent = strings.ReplaceAll(updatedContent,
			"go tool buf",
			bufPath)

		if err := os.WriteFile(generatePath, []byte(updatedContent), 0644); err != nil {
			t.Fatalf("Failed to write updated %s: %v", generatePath, err)
		}
	}

	entDir := filepath.Join(dstDir, "ent")
	defer os.Chdir(originalDir)

	if err := os.Chdir(entDir); err != nil {
		t.Fatalf("Failed to change to ent directory: %v", err)
	}

	cmd := exec.Command("go", "generate", ".")
	cmd.Env = os.Environ()
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go generate failed: %v\nOutput:\n%s", err, output)
	}

	if err := os.Chdir(originalDir); err != nil {
		t.Fatalf("Failed to change back to original directory: %v", err)
	}

	goldenDir := filepath.Join(srcDir, "ent")
	actualDir := filepath.Join(dstDir, "ent")

	compareDirectory(t, exampleDir,
		filepath.Join(goldenDir, "contract"),
		filepath.Join(actualDir, "contract"))

	compareDirectory(t, exampleDir,
		filepath.Join(goldenDir, "gen"),
		filepath.Join(actualDir, "gen"))
}

func compareDirectory(t *testing.T, exampleName, goldenDir, actualDir string) {
	if _, err := os.Stat(goldenDir); os.IsNotExist(err) {
		return
	}

	err := filepath.Walk(goldenDir, func(goldenPath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(goldenDir, goldenPath)
		if err != nil {
			return err
		}

		actualPath := filepath.Join(actualDir, relPath)

		if _, err := os.Stat(actualPath); os.IsNotExist(err) {
			t.Errorf("[%s] Missing generated file: %s", exampleName, relPath)
			return nil
		}

		goldenContent, err := os.ReadFile(goldenPath)
		if err != nil {
			return err
		}

		actualContent, err := os.ReadFile(actualPath)
		if err != nil {
			return err
		}

		if diff := util.Diff(string(goldenContent), string(actualContent)); diff != "" {
			t.Errorf("[%s] File mismatch: %s\n(-expected +actual):\n%s",
				exampleName, relPath, diff)
		}

		return nil
	})

	if err != nil {
		t.Fatalf("Failed to walk golden directory: %v", err)
	}

	err = filepath.Walk(actualDir, func(actualPath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		relPath, err := filepath.Rel(actualDir, actualPath)
		if err != nil {
			return err
		}
		goldenPath := filepath.Join(goldenDir, relPath)
		if _, err := os.Stat(goldenPath); os.IsNotExist(err) {
			t.Errorf("[%s] Unexpected generated file: %s", exampleName, relPath)
		}

		return nil
	})

	if err != nil {
		t.Fatalf("Failed to walk actual directory: %v", err)
	}
}

func copyDir(src, dst string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return err
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() && (entry.Name() == "contract" || entry.Name() == "gen") {
			continue
		}

		if entry.IsDir() {
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}

func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	srcInfo, err := srcFile.Stat()
	if err != nil {
		return err
	}

	dstFile, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, srcInfo.Mode())
	if err != nil {
		return err
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return err
	}

	return nil
}
