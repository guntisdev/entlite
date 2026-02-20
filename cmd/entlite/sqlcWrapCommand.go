package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

func sqlcWrapCommand(args []string) {
	if len(args) < 2 {
		fmt.Fprintf(os.Stderr, "Error: need at least two arguments (input_dir output_dir)\n")
		os.Exit(1)
	}

	// TODO figure out how to pass entity directory
	entityDir := "."
	_, err := loadEntities(entityDir)
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

			content, err := generateWrapperContent(inputFilePath)
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

func generateWrapperContent(inputFilePath string) (string, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, inputFilePath, nil, parser.ParseComments)
	if err != nil {
		return "", fmt.Errorf("failed to parse file: %w", err)
	}

	inputPackageName := node.Name.Name
	absInputDir, _ := filepath.Abs(filepath.Dir(inputFilePath))
	moduleName, workspaceRoot, err := findModuleInfo(absInputDir)
	if err != nil {
		return "", fmt.Errorf("failed to find module info: %w", err)
	}
	relPath, _ := filepath.Rel(workspaceRoot, absInputDir)
	importPath := filepath.Join(moduleName, relPath)
	importPath = filepath.ToSlash(importPath)

	var sb strings.Builder

	packageName := filepath.Base(filepath.Dir(absInputDir))
	sb.WriteString(fmt.Sprintf("package %s\n\n", packageName))

	sb.WriteString("import (\n")
	sb.WriteString(fmt.Sprintf("\t%s \"%s\"\n", inputPackageName, importPath))
	sb.WriteString(")\n\n")

	for _, decl := range node.Decls {
		switch d := decl.(type) {
		case *ast.GenDecl:
			for _, spec := range d.Specs {
				switch s := spec.(type) {
				case *ast.TypeSpec:
					if s.Name.IsExported() {
						sb.WriteString(fmt.Sprintf("type %s = %s.%s\n", s.Name.Name, inputPackageName, s.Name.Name))
					}
				case *ast.ValueSpec:
					for _, name := range s.Names {
						if name.IsExported() {
							if d.Tok == token.CONST {
								sb.WriteString(fmt.Sprintf("const %s = %s.%s\n", name.Name, inputPackageName, name.Name))
							} else {
								sb.WriteString(fmt.Sprintf("var %s = %s.%s\n", name.Name, inputPackageName, name.Name))
							}
						}
					}
				}
			}
		case *ast.FuncDecl:
			if d.Name.IsExported() && d.Recv == nil {
				sb.WriteString(fmt.Sprintf("var %s = %s.%s\n", d.Name.Name, inputPackageName, d.Name.Name))
			}
		}
	}

	return sb.String(), nil
}

// findModuleInfo finds the go.mod file and returns the module name and workspace root
func findModuleInfo(startDir string) (string, string, error) {
	dir := startDir

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
