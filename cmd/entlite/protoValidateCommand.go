package main

import (
	"fmt"
	"os"
	"path"

	protovalidate "github.com/guntisdev/entlite/internal/generator/protoValidate"
)

func protoValidate(args []string) {
	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, "Error: need at least one argument for output directory\n")
		os.Exit(1)
	}

	outputDir := args[0]
	entityDir := "./schema"
	parsedEntities, err := loadEntities(entityDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading entities: %v\n", err)
		os.Exit(1)
	}

	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Output directory does not exist, must be created by buf %v\n", err)
		os.Exit(1)
	}

	entityImports, err := getEntityImports(entityDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading entity imports: %v\n", err)
		os.Exit(1)
	}

	content, err := protovalidate.Generate(parsedEntities, entityImports)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating proto validation: %v\n", err)
		os.Exit(1)
	}

	outputPath := path.Join(outputDir, "proto_validate.go")
	err = os.WriteFile(outputPath, []byte(content), 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing output file %s: %v\n", outputPath, err)
		os.Exit(1)
	}
}
