package util

import (
	"strings"

	"github.com/guntisdev/entlite/internal/schema"
)

func FieldsToStr(fields []string) string {
	var builder strings.Builder
	for _, field := range fields {
		for _, part := range strings.Split(field, "_") {
			if part == "" {
				continue
			}
			builder.WriteString(strings.ToUpper(part[:1]))
			if len(part) > 1 {
				builder.WriteString(part[1:])
			}
		}
	}

	return builder.String()
}

func FiltersToStr(filters []schema.QueryFilter) string {
	var builder strings.Builder
	for _, filter := range filters {
		for _, part := range strings.Split(filter.Field, "_") {
			if part == "" {
				continue
			}
			builder.WriteString(strings.ToUpper(part[:1]))
			if len(part) > 1 {
				builder.WriteString(part[1:])
			}
		}
	}

	return builder.String()
}
