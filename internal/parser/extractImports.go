package parser

import (
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
