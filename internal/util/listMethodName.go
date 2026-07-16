package util

import (
	"fmt"

	"github.com/guntisdev/entlite/internal/schema"
)

func GenListMethodName(query schema.Query, entityName string) string {
	if query.Type == schema.QueryListAll {
		return fmt.Sprintf("ListAll%s", entityName)
	}

	fieldsStr := FieldsToStr(query.Fields)
	byStr := ""
	if fieldsStr != "" {
		byStr = fmt.Sprintf("By%s", fieldsStr)
	}
	filtersStr := FiltersToStr(query.Filters)
	byFilter := ""
	if filtersStr != "" {
		byFilter = fmt.Sprintf("FilterBy%s", filtersStr)
	}
	methodName := fmt.Sprintf("List%s%s%s", entityName, byStr, byFilter)

	return methodName
}

func GenListRpcName(query schema.Query, entityName string) string {
	if query.Type == schema.QueryListAll {
		return "ListAll"
	}

	fieldsStr := FieldsToStr(query.Fields)
	if fieldsStr != "" {
		return fmt.Sprintf("ListBy%s", fieldsStr)
	}
	filtersStr := FiltersToStr(query.Filters)
	if filtersStr != "" {
		return fmt.Sprintf("FilterBy%s", filtersStr)
	}

	return "List"
}
