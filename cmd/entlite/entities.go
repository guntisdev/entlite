package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/guntisdev/entlite/internal/parser"
	"github.com/guntisdev/entlite/internal/schema"
)

// loadEntities discovers and parses entities from the given directory.
// It resolves the absolute path, validates the directory exists, and returns parsed entities.
func loadEntities(entityDir string) ([]schema.Entity, error) {
	dir, err := filepath.Abs(entityDir)
	if err != nil {
		return nil, fmt.Errorf("resolving path %s: %w", entityDir, err)
	}

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil, fmt.Errorf("entity directory does not exist: %s", dir)
	}

	discoveredEntities, err := parser.DiscoverEntities(dir)
	if err != nil {
		return nil, fmt.Errorf("discovering entities: %w", err)
	}

	parsedEntities, err := parser.ParseEntities(discoveredEntities)
	if err != nil {
		return nil, fmt.Errorf("parsing entities: %w", err)
	}

	return parsedEntities, nil
}
