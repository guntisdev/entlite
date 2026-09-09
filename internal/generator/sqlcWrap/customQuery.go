package sqlcwrap

import (
	"go/ast"
	"strings"

	"github.com/guntisdev/entlite/internal/schema"
)

type wrapKind int

const (
	wrapNone wrapKind = iota
	wrapCreate
	wrapCreateBulk
	wrapUpdate
	wrapGet
	wrapList
	wrapDelete
	wrapDeleteAll
)

// sqlc takes the row type from the select, not from the query name
// return type forwarded instead of wrapped.
func (ctx *generationContext) namedEntityWrap(methodName string) (schema.Entity, wrapKind) {
	if _, isDsl := ctx.dslQueries[methodName]; isDsl {
		return schema.Entity{}, wrapNone
	}

	funcDecl, ok := ctx.methods[methodName]
	if !ok {
		return schema.Entity{}, wrapNone
	}

	if name, ok := strings.CutPrefix(methodName, "CreateBulk"); ok {
		if entity, found := ctx.entityMap[name]; found && ctx.createResultFits(funcDecl, entity) {
			return entity, wrapCreateBulk
		}
		return schema.Entity{}, wrapNone
	}

	if name, ok := strings.CutPrefix(methodName, "Create"); ok {
		if entity, found := ctx.entityMap[name]; found && ctx.createResultFits(funcDecl, entity) {
			return entity, wrapCreate
		}
		return schema.Entity{}, wrapNone
	}

	if name, ok := strings.CutPrefix(methodName, "Update"); ok {
		if entity, found := ctx.entityMap[name]; found && ctx.updateResultFits(funcDecl, entity) {
			return entity, wrapUpdate
		}
		return schema.Entity{}, wrapNone
	}

	if strings.HasPrefix(methodName, "Get") {
		if entity, found := ctx.findEntityForGetMethod(methodName); found && returnsModel(funcDecl, entity.Name) {
			return entity, wrapGet
		}
		return schema.Entity{}, wrapNone
	}

	if strings.HasPrefix(methodName, "List") {
		if entity, found := ctx.findEntityForListMethod(methodName); found && returnsModelSlice(funcDecl, entity.Name) {
			return entity, wrapList
		}
		return schema.Entity{}, wrapNone
	}

	if name, ok := strings.CutPrefix(methodName, "DeleteAll"); ok {
		if entity, found := ctx.entityMap[name]; found && returnsErrorOnly(funcDecl) {
			return entity, wrapDeleteAll
		}
		return schema.Entity{}, wrapNone
	}

	if name, ok := strings.CutPrefix(methodName, "Delete"); ok {
		if entity, found := ctx.entityMap[name]; found && returnsErrorOnly(funcDecl) {
			return entity, wrapDelete
		}
		return schema.Entity{}, wrapNone
	}

	return schema.Entity{}, wrapNone
}

// the create wrapper hands sqlc's result back unchanged
func (ctx *generationContext) createResultFits(funcDecl *ast.FuncDecl, entity schema.Entity) bool {
	types := resultTypes(funcDecl)

	if !entity.InsertReturnsID() {
		return len(types) == 1 && types[0] == errorOnlyReturn
	}

	if len(types) != 2 || types[1] != errorOnlyReturn {
		return false
	}

	idType := entity.GetIdField().Type
	if types[0] == string(idType) {
		return true
	}

	return idType == schema.FieldTypeInt && types[0] == "int64" &&
		(ctx.sqlDialect == schema.SQLite || ctx.sqlDialect == schema.MySQL)
}

func (ctx *generationContext) updateResultFits(funcDecl *ast.FuncDecl, entity schema.Entity) bool {
	if ctx.sqlDialect == schema.MySQL {
		// mysql has no RETURNING, the wrapper reads the row back with the get query
		_, hasGet := entity.PrimaryKeyGetQuery()
		return hasGet && returnsErrorOnly(funcDecl)
	}

	return returnsModel(funcDecl, entity.Name)
}

// (Entity, error)
func returnsModel(funcDecl *ast.FuncDecl, entityName string) bool {
	types := resultTypes(funcDecl)

	return len(types) == 2 && types[0] == entityName && types[1] == errorOnlyReturn
}

// ([]Entity, error)
func returnsModelSlice(funcDecl *ast.FuncDecl, entityName string) bool {
	types := resultTypes(funcDecl)

	return len(types) == 2 && types[0] == "[]"+entityName && types[1] == errorOnlyReturn
}

func returnsErrorOnly(funcDecl *ast.FuncDecl) bool {
	types := resultTypes(funcDecl)

	return len(types) == 1 && types[0] == errorOnlyReturn
}

// one entry per returned value, named results included
func resultTypes(funcDecl *ast.FuncDecl) []string {
	if funcDecl.Type.Results == nil {
		return nil
	}

	var types []string
	for _, result := range funcDecl.Type.Results.List {
		count := 1
		if len(result.Names) > 0 {
			count = len(result.Names)
		}
		for i := 0; i < count; i++ {
			types = append(types, formatType(result.Type))
		}
	}

	return types
}
