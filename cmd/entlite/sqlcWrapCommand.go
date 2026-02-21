package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	sqlcwrap "github.com/guntisdev/entlite/internal/generator/sqlcWrap"
)

func sqlcWrapCommand(args []string) {
	if len(args) < 2 {
		fmt.Fprintf(os.Stderr, "Error: need at least two arguments (input_dir output_dir)\n")
		os.Exit(1)
	}

	// TODO figure out how to pass entity directory
	entityDir := "./schema"
	parsedEntities, err := loadEntities(entityDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading entities: %v\n", err)
		os.Exit(1)
	}

	inputDir := args[0]
	outputDir := args[1]

	if _, err := os.Stat(inputDir); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: input directory does not exist: %s\n", inputDir)
		os.Exit(1)
	}

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating output directory: %v\n", err)
		os.Exit(1)
	}

	files, err := os.ReadDir(inputDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input directory: %v\n", err)
		os.Exit(1)
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		fileName := file.Name()
		if strings.HasSuffix(fileName, ".go") {
			fmt.Printf("Processing: %s\n", fileName)

			inputFilePath := filepath.Join(inputDir, fileName)
			outputFilePath := filepath.Join(outputDir, fileName)

			content, err := sqlcwrap.Generate(inputFilePath, parsedEntities)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error generating wrapper content for %s: %v\n", fileName, err)
				os.Exit(1)
			}

			err = os.WriteFile(outputFilePath, []byte(content), 0644)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error writing output file %s: %v\n", outputFilePath, err)
				os.Exit(1)
			}

			fmt.Printf("Created: %s\n", outputFilePath)
		}
	}
}
