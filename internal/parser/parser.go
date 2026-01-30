package parser

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"

	"github.com/guntisdev/entlite/pkg/entlite"
)

type Parser struct {
	discoveredEntities []string
	entityDirectory    string
	parsedEntities     []entlite.Schema
}

func NewParser() *Parser {
	return &Parser{}
}

func (p *Parser) DiscoverEntities(entityDir string) error {
	p.entityDirectory = entityDir
	p.discoveredEntities = nil

	pattern := filepath.Join(entityDir, "*.go")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return fmt.Errorf("failed to fined go files in %s: %w", entityDir, err)
	}

	fmt.Printf("Found %d files\n", len(matches))

	for _, file := range matches {

		entities, err := p.findEntitiesInFile(file)
		if err != nil {
			return fmt.Errorf("failed to examine file %s: %w", file, err)
		}

		p.discoveredEntities = append(p.discoveredEntities, entities...)

		// TODO remove
		for _, entityName := range entities {
			fmt.Printf("Found entities %s\n", entityName)
		}
	}

	return nil
}

func (p *Parser) findEntitiesInFile(filename string) ([]string, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filename, nil, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to parse file %s: %w", filename, err)
	}

	var entities []string

	for _, decl := range file.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.TYPE {
			continue
		}

		for _, spec := range genDecl.Specs {
			typeSec, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}

			sturctType, ok := typeSec.Type.(*ast.StructType)
			if !ok {
				continue
			}

			if p.embedsEntliteSchema(sturctType) {
				entities = append(entities, typeSec.Name.Name)
			}
		}
	}

	return entities, nil
}

func (p *Parser) embedsEntliteSchema(structType *ast.StructType) bool {
	if structType.Fields == nil {
		return false
	}

	for _, field := range structType.Fields.List {
		if len(field.Names) == 0 {
			switch fieldType := field.Type.(type) {
			case *ast.SelectorExpr:
				if ident, ok := fieldType.X.(*ast.Ident); ok && ident.Name == "entlite" {
					if fieldType.Sel.Name == "Schema" {
						return true
					}
				}
			}
		}
	}

	return false
}
