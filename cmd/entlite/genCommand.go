package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/guntisdev/entlite/internal/parser"
)

func genCommand(args []string) {
	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, "Error: entity directory is required")
		os.Exit(1)
	}

	entityDir := args[0]
	dir, err := filepath.Abs(entityDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error resolving path %s: %v\n", entityDir, err)
		os.Exit(1)
	}

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: entity directory: %s does not exist \n", dir)
		os.Exit(1)
	}

	discoveredEntities, err := parser.DiscoverEntities(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error discovering and parsing entities: %v\n", err)
		os.Exit(1)
	}

	parsedEntities, err := parser.ParseEntities(discoveredEntities)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing entities: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Parsed entities:\n%v\n", parsedEntities)
}
