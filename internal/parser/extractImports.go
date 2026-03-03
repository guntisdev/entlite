package parser

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"strings"
)

type ImportInfo struct {
	Name       string // package name or alias
	Path       string // import path
	SchemaPath string // schema file where this import was found
}

func ExtractImports(schemaFilePaths []string) (map[string]ImportInfo, error) {
	pkgToImport := make(map[string]ImportInfo)

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

			// Only store if not already present (first occurrence wins)
			if _, exists := pkgToImport[pkgName]; !exists {
				pkgToImport[pkgName] = ImportInfo{
					Name:       pkgName,
					Path:       importPath,
					SchemaPath: path,
				}
			}
		}
	}

	return pkgToImport, nil
}

func FilterValidateImports(allImports map[string]ImportInfo) map[string]ImportInfo {
	schemaPathsMap := make(map[string]bool)
	for _, importInfo := range allImports {
		schemaPathsMap[importInfo.SchemaPath] = true
	}

	usedPackages := make(map[string]bool)

	for path := range schemaPathsMap {
		fset := token.NewFileSet()
		node, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if err != nil {
			continue
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
						// ident.Name is the package name (e.g., "logic", "env")
						usedPackages[ident.Name] = true
					}
				}
			}

			return true
		})
	}

	result := make(map[string]ImportInfo)
	for pkgName := range usedPackages {
		if importInfo, exists := allImports[pkgName]; exists {
			result[pkgName] = importInfo
		}
	}

	return result
}
