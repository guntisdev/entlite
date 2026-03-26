package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	sqlcwrap "github.com/guntisdev/entlite/internal/generator/sqlcWrap"
	"github.com/guntisdev/entlite/internal/util"
)

func sqlcWrapCommand() {
	entityDir := "./schema"
	parsedEntities, err := loadEntities(entityDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading entities: %v\n", err)
		os.Exit(1)
	}

	entityImports, err := getEntityImports(entityDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading entity imports: %v\n", err)
		os.Exit(1)
	}

	inputDir, err := util.GetSqlcOutputDirFromYaml("./sqlc.yaml")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed reading output directory from sqlc.yaml: %v\n", err)
		os.Exit(1)
	}
	outputDir := filepath.Dir(inputDir)
	pbDir := filepath.Join(filepath.Dir(outputDir), "pb")

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

	dialect, err := util.GetSqlDialectFromSqlcYaml("./sqlc.yaml")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed reading sqlc.yaml: %v\n", err)
		os.Exit(1)
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		fileName := file.Name()
		if strings.HasSuffix(fileName, ".go") {
			inputFilePath := filepath.Join(inputDir, fileName)
			outputFilePath := filepath.Join(outputDir, fileName)

			content, err := sqlcwrap.Generate(inputFilePath, pbDir, parsedEntities, entityImports, dialect)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error generating wrapper content for %s: %v\n", fileName, err)
				os.Exit(1)
			}

			err = os.WriteFile(outputFilePath, []byte(content), 0644)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error writing output file %s: %v\n", outputFilePath, err)
				os.Exit(1)
			}
		}
	}
}
