package docs

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/guntisdev/entlite/internal/util"
)

// Pipeline carries the state from one step to the next.
type Pipeline struct {
	cfg   Config
	files map[string][]byte // step output, keyed by path relative to Root
}

func newPipeline(cfg Config) *Pipeline {
	return &Pipeline{cfg: cfg, files: map[string][]byte{}}
}

// Config returns the settings of the current run.
func (p *Pipeline) Config() Config {
	return p.cfg
}

// Abs turns a path relative to Root into an absolute one.
func (p *Pipeline) Abs(path string) string {
	return filepath.Join(p.cfg.Root, filepath.FromSlash(path))
}

// Out returns the path of a markdown file inside the output directory.
func (p *Pipeline) Out(parts ...string) string {
	return path(append([]string{p.cfg.OutDir}, parts...)...)
}

// HTML returns the path of a file inside the html directory.
func (p *Pipeline) HTML(parts ...string) string {
	return path(append([]string{p.cfg.HTMLDir}, parts...)...)
}

// Read returns a file, an earlier step's output wins over the one on disk.
func (p *Pipeline) Read(relPath string) ([]byte, error) {
	if content, ok := p.files[relPath]; ok {
		return content, nil
	}

	content, err := os.ReadFile(p.Abs(relPath))
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", relPath, err)
	}

	return content, nil
}

// Write records a file, it reaches the disk when every step is done.
func (p *Pipeline) Write(relPath string, content []byte) {
	p.files[relPath] = content
}

// paths returns the recorded paths in a stable order.
func (p *Pipeline) paths() []string {
	paths := make([]string, 0, len(p.files))
	for relPath := range p.files {
		paths = append(paths, relPath)
	}
	sort.Strings(paths)

	return paths
}

// flush writes every recorded file to disk.
func (p *Pipeline) flush() (Result, error) {
	written := p.paths()

	for _, relPath := range written {
		absPath := p.Abs(relPath)
		if err := os.MkdirAll(filepath.Dir(absPath), 0755); err != nil {
			return Result{}, fmt.Errorf("failed to create directory for %s: %w", relPath, err)
		}
		if err := os.WriteFile(absPath, p.files[relPath], 0644); err != nil {
			return Result{}, fmt.Errorf("failed to write %s: %w", relPath, err)
		}
	}

	return Result{Written: written, Stale: nil, Diff: ""}, nil
}

// check compares every recorded file with the tree, it writes nothing.
func (p *Pipeline) check() (Result, error) {
	var stale []string
	var diff strings.Builder

	for _, relPath := range p.paths() {
		generated := string(p.files[relPath])

		onDisk, err := os.ReadFile(p.Abs(relPath))
		if err != nil && !os.IsNotExist(err) {
			return Result{}, fmt.Errorf("failed to read %s: %w", relPath, err)
		}
		if string(onDisk) == generated {
			continue
		}

		stale = append(stale, relPath)
		diff.WriteString(fmt.Sprintf("%s\n%s\n\n", relPath, util.Diff(string(onDisk), generated)))
	}

	return Result{Written: nil, Stale: stale, Diff: diff.String()}, nil
}

// path joins parts into a slash path, empty parts are dropped.
func path(parts ...string) string {
	kept := make([]string, 0, len(parts))
	for _, part := range parts {
		if part != "" {
			kept = append(kept, part)
		}
	}

	return strings.Join(kept, "/")
}
