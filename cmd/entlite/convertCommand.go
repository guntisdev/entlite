package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/guntisdev/entlite/internal/generator/convert"
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

	moduleName, err := getModuleName()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading module name: %v\n", err)
		os.Exit(1)
	}

	dbImport := filepath.Join(moduleName, filepath.Clean(dbDir))
	pbImport := filepath.Join(moduleName, filepath.Clean(pbDir))

	dbImport = filepath.ToSlash(dbImport)
	pbImport = filepath.ToSlash(pbImport)

	imports := []string{
		fmt.Sprintf(`"%s"`, dbImport),
		fmt.Sprintf(`"%s"`, pbImport),
	}

	convertDir := "./convert"
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

	content, err := convert.Generate(parsedEntities, imports)

	if err := writeFile(convertPath, content); err != nil {
		fmt.Fprintf(os.Stderr, "Error failed write to file: %s %v\n", convertPath, err)
		os.Exit(1)
	}
}

func getModuleName() (string, error) {
	file, err := os.Open("go.mod")
	if err != nil {
		return "", fmt.Errorf("failed to open go.mod: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module")), nil
		}
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("error reading go.mod: %w", err)
	}

	return "", fmt.Errorf("module declaration not found in go.mod")
}

func writeFile(filePath, content string) error {
	return os.WriteFile(filePath, []byte(content), 0644)
}
