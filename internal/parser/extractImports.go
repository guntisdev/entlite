package parser

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"strings"
)

// return keyVal pkgName=importPath
func ExtractImports(schemaFilePaths []string) (map[string]string, error) {
	pkgToImport := make(map[string]string)

	for _, path := range schemaFilePaths {
		fset := token.NewFileSet()
		node, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if err != nil {
			continue
		}

		for _, imp := range node.Imports {
			importPath := strings.Trim(imp.Path.Value, `"`)
			var pkgName string

			if imp.Name != nil {
				// Has alias: import foo "github.com/bar/baz"
				pkgName = imp.Name.Name
			} else {
				// No alias, use last part: "github.com/google/uuid" → "uuid"
				pkgName = filepath.Base(importPath)
			}

			pkgToImport[pkgName] = importPath
		}
	}

	return pkgToImport, nil
}

// ExtractValidateImports extracts only imports that are used in .Validate() method calls
// For example: field.String("name").Validate(logic.StartsWithCapital) → returns "logic" import
func ExtractValidateImports(schemaFilePaths []string) (map[string]string, error) {
	allImports := make(map[string]string)
	usedPackages := make(map[string]bool)

	for _, path := range schemaFilePaths {
		fset := token.NewFileSet()
		node, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if err != nil {
			continue
		}

		for _, imp := range node.Imports {
			importPath := strings.Trim(imp.Path.Value, `"`)
			var pkgName string

			if imp.Name != nil {
				pkgName = imp.Name.Name
			} else {
				pkgName = filepath.Base(importPath)
			}

			allImports[pkgName] = importPath
		}

		ast.Inspect(node, func(n ast.Node) bool {
			callExpr, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}

			selExpr, ok := callExpr.Fun.(*ast.SelectorExpr)
			if !ok || selExpr.Sel.Name != "Validate" {
				return true
			}

			// Inspect the arguments to Validate()
			for _, arg := range callExpr.Args {
				if sel, ok := arg.(*ast.SelectorExpr); ok {
					if ident, ok := sel.X.(*ast.Ident); ok {
						// iden.Name is the package name (e.g., "logic", "env")
						usedPackages[ident.Name] = true
					}
				}
			}

			return true
		})
	}

	result := make(map[string]string)
	for pkgName := range usedPackages {
		if importPath, exists := allImports[pkgName]; exists {
			result[pkgName] = importPath
		}
	}

	return result, nil
}
