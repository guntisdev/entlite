package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/guntisdev/entlite/internal/generator/proto"
	"github.com/guntisdev/entlite/internal/generator/sqlc"
)

func genCommand(args []string) {
	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, "Error: entity directory is required")
		os.Exit(1)
	}

	entityDir := args[0]
	parsedEntities, err := loadEntities(entityDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading entities: %v\n", err)
		os.Exit(1)
	}

	dir, _ := filepath.Abs(entityDir)
	protoDir := filepath.Join(filepath.Dir(dir), "contract", "proto")
	sqlcDir := filepath.Join(filepath.Dir(dir), "contract", "sqlc")
	genDir := filepath.Join(filepath.Dir(dir), "gen")

	dirs := []string{protoDir, sqlcDir, genDir}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating directory %s %v\n", dir, err)
			os.Exit(1)
		}
	}

	// PROTO
	if err := proto.Generate(parsedEntities, protoDir); err != nil {
		fmt.Fprintf(os.Stderr, "Failed generating proto: %v\n", err)
		os.Exit(1)
	}

	// TODO get dialect from arguments (if not choose default one)
	sqlcGenerator := sqlc.NewGenerator(sqlc.PostgreSQL)
	if err := sqlcGenerator.Generate(parsedEntities, sqlcDir); err != nil {
		fmt.Fprintf(os.Stderr, "Failed generating sqlc: %v\n", err)
		os.Exit(1)
	}
}
