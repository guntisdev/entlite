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

	dbImport, err := getImportPath(moduleName, dbDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting db import path: %v\n", err)
		os.Exit(1)
	}

	pbImport, err := getImportPath(moduleName, pbDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting pb import path: %v\n", err)
		os.Exit(1)
	}

	imports := []string{
		fmt.Sprintf(`"%s"`, dbImport),
		fmt.Sprintf(`"%s"`, pbImport),
	}

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

	content, err := convert.Generate(parsedEntities, imports)
	if err := writeFile(convertPath, content); err != nil {
		fmt.Fprintf(os.Stderr, "Error failed write to file: %s %v\n", convertPath, err)
		os.Exit(1)
	}
}

func getModuleName() (string, error) {
	goModPath, err := findGoMod()
	if err != nil {
		return "", err
	}

	file, err := os.Open(goModPath)
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

func findGoMod() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get working directory: %w", err)
	}

	for {
		goModPath := filepath.Join(dir, "go.mod")
		if _, err := os.Stat(goModPath); err == nil {
			return goModPath, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("go.mod not found in current directory or any parent directory")
		}
		dir = parent
	}
}

func getImportPath(moduleName, relativePath string) (string, error) {
	// Get current working directory
	_, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get working directory: %w", err)
	}

	// Get absolute path of the target directory
	absPath, err := filepath.Abs(relativePath)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Find the module root (directory containing go.mod)
	goModPath, err := findGoMod()
	if err != nil {
		return "", err
	}
	moduleRoot := filepath.Dir(goModPath)

	// Get relative path from module root to target
	relPath, err := filepath.Rel(moduleRoot, absPath)
	if err != nil {
		return "", fmt.Errorf("failed to get relative path from module root: %w", err)
	}

	// If the path starts with "..", it's outside the module
	if strings.HasPrefix(relPath, "..") {
		return "", fmt.Errorf("path %s is outside module root %s", absPath, moduleRoot)
	}

	// Join module name with relative path and convert to forward slashes
	importPath := filepath.Join(moduleName, relPath)
	importPath = filepath.ToSlash(importPath)

	return importPath, nil
}

func writeFile(filePath, content string) error {
	return os.WriteFile(filePath, []byte(content), 0644)
}
