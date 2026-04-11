package parser

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"

	"github.com/guntisdev/entlite/internal/schema"
)

func ParseEntities(discoveredEntities []DiscoveredEntity) ([]schema.Entity, error) {
	var entities []schema.Entity

	for _, discovered := range discoveredEntities {
		parsed, err := parseEntityFromFile(discovered)
		if err != nil {
			return nil, fmt.Errorf("entity %q in %s: %w", discovered.Name, discovered.Path, err)
		}
		entities = append(entities, parsed)
	}

	return entities, nil
}

func parseEntityFromFile(discovered DiscoveredEntity) (schema.Entity, error) {
	entity := schema.Entity{
		Name: discovered.Name,
	}

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, discovered.Path, nil, parser.ParseComments)
	if err != nil {
		return entity, fmt.Errorf("failed to parse file %s: %w", discovered.Path, err)
	}

	for _, decl := range file.Decls {
		funcDecl, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}

		if funcDecl.Recv == nil || len(funcDecl.Recv.List) == 0 {
			continue
		}

		recvType := funcDecl.Recv.List[0].Type
		var recvTypeName string

		switch t := recvType.(type) {
		case *ast.Ident:
			recvTypeName = t.Name
		case *ast.StarExpr:
			if ident, ok := t.X.(*ast.Ident); ok {
				recvTypeName = ident.Name
			}
		}

		if recvTypeName != entity.Name {
			continue
		}

		// Parse Annotations
		if funcDecl.Name.Name == "Annotations" {
			annotations, err := parseAnnotationsMethod(funcDecl)
			if err != nil {
				return entity, fmt.Errorf("failed to parse annotations: %w", err)
			}
			entity.Annotations = annotations
		}

		// Parse Fields
		if funcDecl.Name.Name == "Fields" {
			fields, err := parseFieldsMethod(funcDecl)
			if err != nil {
				return entity, fmt.Errorf("failed to parse fields: %w", err)
			}

			if err := checkProtoFieldCollision(fields); err != nil {
				return entity, err
			}

			// add protoField, add id if not there
			fields = addFieldNumbers(fields)
			entity.Fields = fields
		}

		// Parse Queries
		if funcDecl.Name.Name == "Queries" {
			queries, err := parseQueriesMethod(funcDecl)
			if err != nil {
				return entity, fmt.Errorf("failed to parse queries: %w", err)
			}
			entity.Queries = queries
		}
	}

	if err := validateQueryFields(entity); err != nil {
		return entity, err
	}

	return entity, nil
}
