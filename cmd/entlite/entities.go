package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/guntisdev/entlite/internal/parser"
	"github.com/guntisdev/entlite/internal/schema"
)

// loadEntities discovers and parses entities from the given directory
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

func getEntityImports(entityDir string) (map[string]string, error) {
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

	schemaFilePaths := make([]string, len(discoveredEntities))
	for i, discovered := range discoveredEntities {
		schemaFilePaths[i] = discovered.Path
	}

	entityImports, err := parser.ExtractImports(schemaFilePaths)
	if err != nil {
		return nil, fmt.Errorf("extracting imports: %w", err)
	}

	// Filter out entlite DSL imports - we only need actual dependencies like time, custom logic, etc.
	filteredImports := make(map[string]string)
	for alias, importPath := range entityImports {
		if !strings.HasPrefix(importPath, "github.com/guntisdev/entlite/pkg/entlite") {
			filteredImports[alias] = importPath
		}
	}

	return filteredImports, nil
}

func getValidateImports(entityDir string) (map[string]string, error) {
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

	schemaFilePaths := make([]string, len(discoveredEntities))
	for i, discovered := range discoveredEntities {
		schemaFilePaths[i] = discovered.Path
	}

	validateImports, err := parser.ExtractValidateImports(schemaFilePaths)
	if err != nil {
		return nil, fmt.Errorf("extracting validate imports: %w", err)
	}

	return validateImports, nil
}
