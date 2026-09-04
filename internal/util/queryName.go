package util

import (
	"fmt"
	"strings"

	"github.com/guntisdev/entlite/internal/schema"
)

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
		return genListName(query, entityName)
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

func GenEntityGetByPrimaryKeyName(entity schema.Entity) string {
	keyFields := entity.PrimaryKeyFields()
	names := make([]string, 0, len(keyFields))
	for _, field := range keyFields {
		names = append(names, field.Name)
	}

	for _, query := range entity.Queries {
		if query.Type != schema.QueryGetBy || len(query.Fields) != len(names) {
			continue
		}
		if !sameFields(query.Fields, names) {
			continue
		}

		return GenQueryName(query, entity.Name)
	}

	return GenQueryName(schema.Query{Type: schema.QueryGetBy, Fields: names}, entity.Name)
}

func sameFields(a, b []string) bool {
	for i := range a {
		if !strings.EqualFold(a[i], b[i]) {
			return false
		}
	}

	return true
}

func genListName(query schema.Query, entityName string) string {
	if query.Type == schema.QueryListAll {
		return fmt.Sprintf("ListAll%s", entityName)
	}

	byStr := ""
	if fieldsStr := FieldsToStr(query.Fields); fieldsStr != "" {
		byStr = fmt.Sprintf("By%s", fieldsStr)
	}

	byFilter := ""
	if filtersStr := FiltersToStr(query.Filters); filtersStr != "" {
		byFilter = fmt.Sprintf("FilterBy%s", filtersStr)
	}

	return fmt.Sprintf("List%s%s%s", entityName, byStr, byFilter)
}
