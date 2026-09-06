package parser

import (
	"fmt"
	"strings"

	"github.com/guntisdev/entlite/internal/schema"
)

// names every query that has no custom Name()
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
		return fmt.Sprintf("Get%sBy%s", entityName, fieldsToStr(query.Fields))
	case schema.QueryListBy, schema.QueryListAll:
		return genListName(query, entityName)
	default:
		return ""
	}
}

func genListName(query schema.Query, entityName string) string {
	if query.Type == schema.QueryListAll {
		return fmt.Sprintf("ListAll%s", entityName)
	}

	byStr := ""
	if fieldsStr := fieldsToStr(query.Fields); fieldsStr != "" {
		byStr = fmt.Sprintf("By%s", fieldsStr)
	}

	byFilter := ""
	if filtersStr := filtersToStr(query.Filters); filtersStr != "" {
		byFilter = fmt.Sprintf("FilterBy%s", filtersStr)
	}

	return fmt.Sprintf("List%s%s%s", entityName, byStr, byFilter)
}

func fieldsToStr(fields []string) string {
	var builder strings.Builder
	for _, field := range fields {
		builder.WriteString(toCamelCase(field))
	}

	return builder.String()
}

func filtersToStr(filters []schema.QueryFilter) string {
	var builder strings.Builder
	for _, filter := range filters {
		builder.WriteString(toCamelCase(filter.Field))
	}

	return builder.String()
}

func toCamelCase(field string) string {
	var builder strings.Builder
	for _, part := range strings.Split(field, "_") {
		if part == "" {
			continue
		}
		builder.WriteString(strings.ToUpper(part[:1]))
		if len(part) > 1 {
			builder.WriteString(part[1:])
		}
	}

	return builder.String()
}
