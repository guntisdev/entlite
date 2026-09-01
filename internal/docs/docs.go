// Package docs builds the entlite documentation from the source tree.
package docs

import (
	"fmt"
	"os"
)

// Config tells the pipeline where to read from and where to write to.
type Config struct {
	// Root is the repo root, it defaults to the working directory.
	Root string
	// OutDir holds the generated markdown, relative to Root.
	OutDir string
	// HTMLDir holds the generated html, empty skips the site step.
	HTMLDir string
	// Check reports stale files instead of writing them.
	Check bool
}

// Step is one stage of the pipeline.
type Step struct {
	// Name is used in log and error messages.
	Name string
	// HTMLOnly marks a step that only runs when HTMLDir is set.
	HTMLOnly bool
	// Run fills the pipeline with the files the step produces.
	Run func(p *Pipeline) error
}

// steps run in this order. readme injection comes before the example pages,
// so they read the final readme and not a stale table
var steps = []Step{
	{Name: "reference", HTMLOnly: false, Run: reference},
	{Name: "readme", HTMLOnly: false, Run: readme},
	{Name: "examples", HTMLOnly: false, Run: examples},
	{Name: "site", HTMLOnly: true, Run: site},
}

// Result reports what a run produced.
type Result struct {
	// Written lists the paths written to disk, empty in check mode.
	Written []string
	// Stale lists the paths that differ from the tree, only filled in check mode.
	Stale []string
	// Diff is a readable diff of the stale files, only filled in check mode.
	Diff string
}

// Run executes every step and then writes or checks the result.
func Run(cfg Config) (Result, error) {
	if cfg.Root == "" {
		wd, err := os.Getwd()
		if err != nil {
			return Result{}, fmt.Errorf("failed to find working directory: %w", err)
		}
		cfg.Root = wd
	}
	if cfg.OutDir == "" {
		cfg.OutDir = "docs"
	}

	p := newPipeline(cfg)

	for _, step := range steps {
		if step.HTMLOnly && (cfg.HTMLDir == "" || cfg.Check) {
			continue
		}
		if err := step.Run(p); err != nil {
			return Result{}, fmt.Errorf("step %s failed: %w", step.Name, err)
		}
	}

	if cfg.Check {
		return p.check()
	}

	return p.flush()
}
