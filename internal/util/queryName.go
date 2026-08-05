package util

import (
	"fmt"

	"github.com/guntisdev/entlite/internal/schema"
)

// GenQueryRpcName returns the rpc name of a query inside its entity service.
// A custom Name() from the schema replaces the generated name.
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

// GenQueryMessageName returns the base name of a query's request/response
// messages. A custom Name() from the schema replaces the generated name.
func GenQueryMessageName(query schema.Query, entityName string) string {
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

// GenEntityQueryMessageName returns the base message name of the entity's first
// query of the given type, falling back to the generated name when the entity
// has no such query.
func GenEntityQueryMessageName(entity schema.Entity, queryType schema.QueryType) string {
	for _, query := range entity.Queries {
		if query.Type == queryType {
			return GenQueryMessageName(query, entity.Name)
		}
	}

	return GenQueryMessageName(schema.Query{Type: queryType}, entity.Name)
}
