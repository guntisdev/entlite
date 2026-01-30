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

type DiscoveredEntity struct {
	Name string
	Path string
}

func NewParser() *Parser {
	return &Parser{}
}

func (p *Parser) DiscoverEntities(entityDir string) ([]DiscoveredEntity, error) {
	p.entityDirectory = entityDir
	p.discoveredEntities = nil

	pattern := filepath.Join(entityDir, "*.go")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to fined go files in %s: %w", entityDir, err)
	}

	fmt.Printf("Found %d files\n", len(matches))

	var entityList []DiscoveredEntity

	for _, file := range matches {

		entities, err := p.findEntitiesInFile(file)
		if err != nil {
			return nil, fmt.Errorf("failed to examine file %s: %w", file, err)
		}

		for _, entity := range entities {
			entityList = append(entityList, DiscoveredEntity{file, entity})
		}

		// TODO remove
		for _, entityName := range entities {
			fmt.Printf("Found entities %s\n", entityName)
		}
	}

	if len(entityList) == 0 {
		return nil, fmt.Errorf("no entities found in directory")
	}

	return entityList, nil
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
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}

			structType, ok := typeSpec.Type.(*ast.StructType)
			if !ok {
				continue
			}

			if p.embedsEntliteSchema(structType) {
				entities = append(entities, typeSpec.Name.Name)
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
