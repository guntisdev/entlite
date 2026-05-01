package util

import "strings"

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
