package util

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// startDir accepts both relative and absolute path
// rootDir is absolute path for go.mod
func FindModuleInfo(startDir string) (moduleName string, rootDir string, err error) {
	dir, err := filepath.Abs(startDir)
	if err != nil {
		return "", "", fmt.Errorf("failed to get absolute path: %w", err)
	}

	for {
		goModPath := filepath.Join(dir, "go.mod")
		if _, err := os.Stat(goModPath); err == nil {
			content, err := os.ReadFile(goModPath)
			if err != nil {
				return "", "", err
			}

			lines := strings.Split(string(content), "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if strings.HasPrefix(line, "module ") {
					moduleName := strings.TrimSpace(strings.TrimPrefix(line, "module"))
					return moduleName, dir, nil
				}
			}
			return "", "", fmt.Errorf("module declaration not found in go.mod")
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return "", "", fmt.Errorf("go.mod not found")
}

func PathToImport(path string) (string, error) {
	absolutPath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("Getting absolute path: %w", err)
	}

	info, err := os.Stat(absolutPath)
	if err == nil && !info.IsDir() {
		absolutPath = filepath.Dir(absolutPath)
	}

	moduleName, moduleRoot, err := FindModuleInfo(absolutPath)
	if err != nil {
		return "", fmt.Errorf("Finding module info: %w", err)
	}

	relPath, err := filepath.Rel(moduleRoot, absolutPath)
	if err != nil {
		return "", fmt.Errorf("Getting relative path: %w", err)
	}

	if strings.HasPrefix(relPath, "..") {
		return "", fmt.Errorf("path %s is outside module root %s", absolutPath, moduleRoot)
	}

	importPath := filepath.Join(moduleName, relPath)
	importPath = filepath.ToSlash(importPath)

	return importPath, nil
}
