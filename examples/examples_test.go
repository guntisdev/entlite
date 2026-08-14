//go:build integration
// +build integration

// Integration tests that generate each example in place and check the result
// against what is committed.
package examples

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Directories owned by codegen, relative to the example.
var generated = []string{"ent/contract", "ent/gen"}

type example struct {
	dir  string // path relative to examples/, slash separated
	name string // subtest name, no slashes to keep -run patterns simple
}

func TestExamples(t *testing.T) {
	examples, err := findExamples(".")
	if err != nil {
		t.Fatalf("Failed to discover examples: %v", err)
	}
	if len(examples) == 0 {
		t.Fatal("No examples found")
	}

	for _, ex := range examples {
		t.Run(ex.name, func(t *testing.T) {
			t.Parallel()

			if !t.Run("codegen", func(t *testing.T) { testCodegen(t, ex) }) {
				return
			}

			// Owned here because the binary outlives the build step
			binDir, bin := t.TempDir(), ""
			if !t.Run("build", func(t *testing.T) { bin = testBuild(t, ex, binDir) }) {
				return
			}

			t.Run("web", func(t *testing.T) { testWeb(t, ex) })
			t.Run("run", func(t *testing.T) { testRun(t, ex, bin) })
		})
	}
}

// findExamples returns every directory below root that holds an ent/generate.go.
// Examples are not searched for nested examples.
func findExamples(root string) ([]example, error) {
	var found []example

	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			return nil
		}
		if name := d.Name(); name == "node_modules" || (strings.HasPrefix(name, ".") && path != root) {
			return filepath.SkipDir
		}

		if _, err := os.Stat(filepath.Join(path, "ent", "generate.go")); err != nil {
			return nil
		}

		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		dir := filepath.ToSlash(rel)
		found = append(found, example{dir: dir, name: strings.ReplaceAll(dir, "/", "_")})

		return filepath.SkipDir
	})

	return found, err
}

// testCodegen regenerates the example and fails if the output differs from git.
func testCodegen(t *testing.T, ex example) {
	requireClean(t, ex)

	entDir := filepath.Join(filepath.FromSlash(ex.dir), "ent")

	// buf dep update rewrites the lock file, that is not a codegen result
	keep(t, filepath.Join(entDir, "buf.lock"))

	if out, err := run(entDir, "go", "generate", "."); err != nil {
		t.Fatalf("go generate failed: %v\nOutput:\n%s", err, out)
	}

	dirty := status(t, ex, generated...)
	if dirty == "" {
		return
	}

	diff, _ := run(".", append([]string{"git", "diff", "--"}, paths(ex, generated...)...)...)
	t.Errorf("Generated output differs from what is committed.\n"+
		"Changed files:\n%s\nDiff:\n%s\n"+
		"Run `cd examples/%s/ent && go generate .` and commit the result if this is expected.",
		dirty, diff, ex.dir)
}

// testBuild compiles the example and returns its server binary.
func testBuild(t *testing.T, ex example, binDir string) string {
	if out, err := run(filepath.FromSlash(ex.dir), "go", "build", "-o", binDir+string(os.PathSeparator), "./..."); err != nil {
		t.Fatalf("go build failed: %v\nOutput:\n%s", err, out)
	}

	entries, err := os.ReadDir(binDir)
	if err != nil {
		t.Fatalf("Failed to read build output: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("Expected exactly one binary in %s, got %d", binDir, len(entries))
	}

	return filepath.Join(binDir, entries[0].Name())
}

// testWeb bundles the frontend, which fails when the generated client no longer
// fits the example code.
func testWeb(t *testing.T, ex example) {
	webDir := filepath.Join(filepath.FromSlash(ex.dir), "web")
	if _, err := os.Stat(filepath.Join(webDir, "package.json")); err != nil {
		t.Skip("No web/package.json")
	}
	if _, err := exec.LookPath("npm"); err != nil {
		t.Skip("npm is not installed")
	}

	if _, err := os.Stat(filepath.Join(webDir, "node_modules")); err != nil {
		if out, err := run(webDir, "npm", "ci", "--no-audit", "--no-fund"); err != nil {
			t.Fatalf("npm ci failed: %v\nOutput:\n%s", err, out)
		}
	}

	if out, err := run(webDir, "npm", "run", "build"); err != nil {
		t.Errorf("npm run build failed: %v\nOutput:\n%s", err, out)
	}
}

// testRun starts the server and checks /health, to catch a crash on startup.
func testRun(t *testing.T, ex example, bin string) {
	exDir := filepath.FromSlash(ex.dir)
	if _, err := os.Stat(filepath.Join(exDir, "docker-compose.yml")); err == nil {
		t.Skip("Needs a database from docker compose")
	}

	// The example creates this database, remove it if it was not there before
	dbPath := filepath.Join(exDir, "server", "db.db")
	if _, err := os.Stat(dbPath); errors.Is(err, os.ErrNotExist) {
		t.Cleanup(func() { os.Remove(dbPath) })
	}

	port := freePort(t)
	cmd := exec.Command(bin, fmt.Sprintf("-port=%d", port))
	cmd.Dir = exDir
	var output strings.Builder
	cmd.Stdout = &output
	cmd.Stderr = &output

	if err := cmd.Start(); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}

	exited := make(chan error, 1)
	go func() { exited <- cmd.Wait() }()

	t.Cleanup(func() {
		_ = cmd.Process.Signal(os.Interrupt)
		select {
		case <-exited:
		case <-time.After(5 * time.Second):
			_ = cmd.Process.Kill()
		}
	})

	healthURL := fmt.Sprintf("http://127.0.0.1:%d/health", port)
	deadline := time.After(30 * time.Second)

	for {
		select {
		case err := <-exited:
			t.Fatalf("Server exited before serving /health: %v\nOutput:\n%s", err, output.String())
		case <-deadline:
			t.Fatalf("Server did not answer %s within 30s\nOutput:\n%s", healthURL, output.String())
		case <-time.After(100 * time.Millisecond):
			resp, err := http.Get(healthURL)
			if err != nil {
				continue
			}
			resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				t.Fatalf("GET /health returned %s\nOutput:\n%s", resp.Status, output.String())
			}
			return
		}
	}
}

// requireClean fails on local changes, which would make the diff after
// generating meaningless.
func requireClean(t *testing.T, ex example) {
	if dirty := status(t, ex, generated...); dirty != "" {
		t.Fatalf("Generated files have uncommitted changes, commit or stash "+
			"them first:\n%s", dirty)
	}
}

// status returns the git porcelain output for the given paths of an example.
func status(t *testing.T, ex example, dirs ...string) string {
	out, err := run(".", append([]string{"git", "status", "--porcelain", "--"}, paths(ex, dirs...)...)...)
	if err != nil {
		t.Fatalf("git status failed: %v\nOutput:\n%s", err, out)
	}

	return out
}

// keep writes a file's content back when the test ends. Restoring by hand
// instead of by git, so the test can never throw away local work.
func keep(t *testing.T, path string) {
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read %s: %v", path, err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Failed to stat %s: %v", path, err)
	}

	t.Cleanup(func() {
		after, err := os.ReadFile(path)
		if err == nil && string(after) == string(before) {
			return
		}
		if err := os.WriteFile(path, before, info.Mode()); err != nil {
			t.Errorf("Failed to restore %s: %v", path, err)
		}
	})
}

func paths(ex example, dirs ...string) []string {
	out := make([]string, 0, len(dirs))
	for _, dir := range dirs {
		out = append(out, filepath.FromSlash(ex.dir+"/"+dir))
	}

	return out
}

func run(dir string, args ...string) (string, error) {
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Dir = dir
	cmd.Env = os.Environ()
	out, err := cmd.CombinedOutput()

	return string(out), err
}

func freePort(t *testing.T) int {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to find a free port: %v", err)
	}
	defer listener.Close()

	return listener.Addr().(*net.TCPAddr).Port
}
