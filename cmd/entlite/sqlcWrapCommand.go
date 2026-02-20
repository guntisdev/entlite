package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func sqlcWrapCommand(args []string) {
	if len(args) < 2 {
		fmt.Fprintf(os.Stderr, "Error: need at least two arguments (input_dir output_dir)\n")
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

			outputFilePath := filepath.Join(outputDir, fileName)
			outputFile, err := os.Create(outputFilePath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error creating output file %s: %v\n", outputFilePath, err)
				os.Exit(1)
			}
			outputFile.Close()
		}
	}
}
