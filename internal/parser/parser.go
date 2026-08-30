package parser

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"

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

	comments := newCommentLookup(fset, file)
	entity.Comment = parseEntityComment(file, entity.Name)

	hasContractsMethod := false

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

		// Parse Contracts
		if funcDecl.Name.Name == "Contracts" {
			hasContractsMethod = true

			contracts, err := parseContractsMethod(funcDecl)
			if err != nil {
				return entity, fmt.Errorf("failed to parse contracts: %w", err)
			}
			entity.Contracts = contracts
		}

		// Parse Fields
		if funcDecl.Name.Name == "Fields" {
			fields, err := parseFieldsMethod(funcDecl, comments)
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
			queries, err := parseQueriesMethod(funcDecl, comments)
			if err != nil {
				return entity, fmt.Errorf("failed to parse queries: %w", err)
			}
			entity.Queries = queries
		}

		// Parse Indexes
		if funcDecl.Name.Name == "Indexes" {
			indexes, err := parseIndexesMethod(funcDecl)
			if err != nil {
				return entity, fmt.Errorf("failed to parse indexes: %w", err)
			}
			entity.Indexes = indexes
		}
	}

	if err := validateContracts(entity, hasContractsMethod); err != nil {
		return entity, err
	}

	// an explicit index.Primary overrides the id field, the compound key becomes the
	// table's only PRIMARY KEY
	applyPrimaryIndexOverride(&entity)

	fields, err := applyFieldContracts(entity)
	if err != nil {
		return entity, err
	}
	entity.Fields = fields

	queries, err := applyQueryContracts(entity)
	if err != nil {
		return entity, err
	}
	entity.Queries = queries

	if err := validateQueryFields(entity); err != nil {
		return entity, err
	}

	if err := validateVirtualFields(entity); err != nil {
		return entity, err
	}

	if err := validateIndexFields(entity); err != nil {
		return entity, err
	}

	if err := validateJSONDefaults(entity); err != nil {
		return entity, err
	}

	return entity, nil
}

// contracts are always explicit, an entity without them generates nothing
func validateContracts(entity schema.Entity, hasMethod bool) error {
	if !hasMethod {
		return fmt.Errorf("entity %q is missing Contracts() method, declare entlite.SQLC() and/or entlite.PROTO()", entity.Name)
	}

	if len(entity.Contracts) == 0 {
		return fmt.Errorf("entity %q has empty Contracts(), declare entlite.SQLC() and/or entlite.PROTO()", entity.Name)
	}

	return nil
}

// catch malformed text at generation time
func validateJSONDefaults(entity schema.Entity) error {
	for _, field := range entity.Fields {
		if field.Type != schema.FieldTypeJSON || field.DefaultValue == nil {
			continue
		}
		text, ok := field.DefaultValue.(string)
		if !ok {
			return fmt.Errorf("entity %q field %q json default must be a string", entity.Name, field.Name)
		}
		if !json.Valid([]byte(text)) {
			return fmt.Errorf("entity %q field %q json default is not valid json: %s", entity.Name, field.Name, text)
		}
	}

	return nil
}

func validateVirtualFields(entity schema.Entity) error {
	for _, field := range entity.Fields {
		if field.IsID() && entity.IsFieldVirtual(field) {
			return fmt.Errorf("entity %q id field cannot be virtual", entity.Name)
		}
	}

	return nil
}

func validateIndexFields(entity schema.Entity) error {
	for _, idx := range entity.Indexes {
		for _, column := range idx.Columns {
			if !entityHasField(entity, column.Name) {
				return fmt.Errorf("entity %q index references nonexisting field %q", entity.Name, column.Name)
			}
			if entityFieldIsVirtual(entity, column.Name) {
				return fmt.Errorf("entity %q index references virtual field %q, which has no database column", entity.Name, column.Name)
			}
			// TODO support indexing json fields (postgres needs GIN, mysql needs a generated column)
			if entityFieldHasType(entity, column.Name, schema.FieldTypeJSON) {
				return fmt.Errorf("entity %q index references json field %q, indexing json fields is not supported", entity.Name, column.Name)
			}
		}
	}

	return nil
}

func applyPrimaryIndexOverride(entity *schema.Entity) {
	hasPrimaryIndex := false
	for _, idx := range entity.Indexes {
		if idx.Type == schema.IndexPrimary {
			hasPrimaryIndex = true
			break
		}
	}
	if !hasPrimaryIndex {
		return
	}

	for i := range entity.Fields {
		if entity.Fields[i].IsID() {
			entity.Fields[i].Primary = false
		}
	}
}

// parseEntityComment reads the doc comment line above the entity type declaration.
func parseEntityComment(file *ast.File, name string) string {
	for _, decl := range file.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.TYPE {
			continue
		}

		for _, spec := range genDecl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok || typeSpec.Name.Name != name {
				continue
			}

			doc := typeSpec.Doc
			if doc == nil {
				doc = genDecl.Doc
			}
			if doc != nil {
				return strings.TrimSpace(doc.Text())
			}
		}
	}

	return ""
}
