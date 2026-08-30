package util

import (
	"fmt"
	"strings"

	"github.com/guntisdev/entlite/internal/schema"
)

// GenQueryName returns the entity qualified query name, used as the sqlc query name
// and the base name of its proto messages. A custom Name() replaces it.
func GenQueryName(query schema.Query, entityName string) string {
	if query.Name != "" {
		return query.Name
	}

	switch query.Type {
	case schema.QueryCreate:
		return fmt.Sprintf("Create%s", entityName)
	case schema.QueryCreateBulk:
		return fmt.Sprintf("CreateBulk%s", entityName)
	case schema.QueryUpdate:
		return fmt.Sprintf("Update%s", entityName)
	case schema.QueryDelete:
		return fmt.Sprintf("Delete%s", entityName)
	case schema.QueryDeleteAll:
		return fmt.Sprintf("DeleteAll%s", entityName)
	case schema.QueryGetBy:
		return fmt.Sprintf("Get%sBy%s", entityName, FieldsToStr(query.Fields))
	case schema.QueryListBy, schema.QueryListAll:
		return GenListMethodName(query, entityName)
	default:
		return ""
	}
}

// GenQueryRpcName returns the rpc name inside the entity service. A custom Name()
// replaces it.
func GenQueryRpcName(query schema.Query, entityName string) string {
	if query.Name != "" {
		return query.Name
	}

	switch query.Type {
	case schema.QueryCreate:
		return "Create"
	case schema.QueryCreateBulk:
		return "CreateBulk"
	case schema.QueryUpdate:
		return "Update"
	case schema.QueryDelete:
		return "Delete"
	case schema.QueryDeleteAll:
		return "DeleteAll"
	case schema.QueryGetBy:
		return fmt.Sprintf("GetBy%s", FieldsToStr(query.Fields))
	case schema.QueryListBy, schema.QueryListAll:
		return GenListRpcName(query, entityName)
	default:
		return ""
	}
}

// GenEntityQueryName returns the entity's first query of the given type, falling back
// to the generated name. For the single-per-entity types: create, update, delete.
func GenEntityQueryName(entity schema.Entity, queryType schema.QueryType) string {
	for _, query := range entity.Queries {
		if query.Type == queryType {
			return GenQueryName(query, entity.Name)
		}
	}

	return GenQueryName(schema.Query{Type: queryType}, entity.Name)
}

// GenEntityGetByIdName returns the name of the query that gets the entity by id.
func GenEntityGetByIdName(entity schema.Entity) string {
	idQuery := schema.Query{Type: schema.QueryGetBy, Fields: []string{"ID"}}

	for _, query := range entity.Queries {
		if query.Type != schema.QueryGetBy || len(query.Fields) != 1 {
			continue
		}
		if !strings.EqualFold(query.Fields[0], "id") {
			continue
		}

		return GenQueryName(query, entity.Name)
	}

	return GenQueryName(idQuery, entity.Name)
}
