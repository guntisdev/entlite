package parser

import (
	"github.com/guntisdev/entlite/internal/naming"
	"github.com/guntisdev/entlite/internal/schema"
)

// name every query that has no custom Name()
func resolveQueryNames(entity *schema.Entity) {
	for i := range entity.Queries {
		if entity.Queries[i].Name != "" {
			continue
		}
		entity.Queries[i].Name = genQueryName(entity.Queries[i], entity.Name)
	}
}

func genQueryName(query schema.Query, entityName string) string {
	switch query.Type {
	case schema.QueryCreate:
		return naming.CreateQueryName(entityName)
	case schema.QueryCreateBulk:
		return naming.CreateBulkQueryName(entityName)
	case schema.QueryUpdate:
		return naming.UpdateQueryName(entityName)
	case schema.QueryDelete:
		return naming.DeleteQueryName(entityName)
	case schema.QueryDeleteAll:
		return naming.DeleteAllQueryName(entityName)
	case schema.QueryGetBy:
		return naming.GetByQueryName(entityName, query.Fields)
	case schema.QueryListAll:
		return naming.ListAllQueryName(entityName, query.Distinct)
	case schema.QueryListBy:
		return naming.ListByQueryName(entityName, query.Distinct, query.Fields, filterFields(query.Filters))
	default:
		return ""
	}
}

func filterFields(filters []schema.QueryFilter) []string {
	fields := make([]string, 0, len(filters))
	for _, filter := range filters {
		fields = append(fields, filter.Field)
	}

	return fields
}
