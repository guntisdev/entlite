package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/guntisdev/entlite/internal/generator/convert"
	"github.com/guntisdev/entlite/internal/util"
)

func convertCommand(args []string) {
	if len(args) < 2 {
		fmt.Fprintf(os.Stderr, "Error: sqlc and proto export directories needed")
	}

	dbDir := args[0]
	pbDir := args[1]

	if _, err := os.Stat(dbDir); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: db directory does not exist: %s\n", dbDir)
		os.Exit(1)
	}

	if _, err := os.Stat(pbDir); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: pb directory does not exist: %s\n", pbDir)
		os.Exit(1)
	}

	dbImport, err := util.PathToImport(dbDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Getting db import path: %v\n", err)
		os.Exit(1)
	}

	pbImport, err := util.PathToImport(pbDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Getting pb import path: %v\n", err)
		os.Exit(1)
	}

	imports := []string{dbImport, pbImport}

	convertDir := "./gen/convert"
	if err := os.MkdirAll(convertDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating convert directory: %v\n", err)
		os.Exit(1)
	}
	convertPath := filepath.Join(convertDir, "convert.go")

	entityDir := "./schema"
	parsedEntities, err := loadEntities(entityDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading entities: %v\n", err)
		os.Exit(1)
	}

	dialect, err := util.GetSqlDialectFromSqlcYaml("./sqlc.yaml")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed reading sqlc.yaml: %v\n", err)
		os.Exit(1)
	}

	content, err := convert.Generate(parsedEntities, imports, dialect)
	if err := writeFile(convertPath, content); err != nil {
		fmt.Fprintf(os.Stderr, "Error failed write to file: %s %v\n", convertPath, err)
		os.Exit(1)
	}
}

func writeFile(filePath, content string) error {
	return os.WriteFile(filePath, []byte(content), 0644)
}
