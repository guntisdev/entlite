package protovalidate

import (
	"fmt"
	"strings"

	"github.com/guntisdev/entlite/internal/schema"
)

func Generate(entities []schema.Entity, imports map[string]string) (string, error) {
	var content strings.Builder

	content.WriteString("package pb\n\n")

	fmtImportNeeded := false
	for _, entity := range entities {
		if hasValidateField(entity) {
			fmtImportNeeded = true
		}
	}

	content.WriteString("import (\n")
	if fmtImportNeeded {
		content.WriteString("\t\"fmt\"\n")
	}

	// TODO this brings all imports, need only validate imports
	for _, importPath := range imports {
		content.WriteString(fmt.Sprintf("\t\"%s\"\n", importPath))
	}
	content.WriteString(")\n")

	for _, entity := range entities {
		if !hasValidateField(entity) {
			continue
		}
		content.WriteString(generateCreateRequest(entity))
		content.WriteString(generateUpdateRequest(entity))
	}

	return content.String(), nil
}

func generateCreateRequest(entity schema.Entity) string {
	var content strings.Builder
	content.WriteString(fmt.Sprintf("func (r *Create%sRequest) Validate() error {\n", entity.Name))
	for _, field := range entity.Fields {
		if field.Validate == nil {
			continue
		}

		fieldName := toProtoFieldName(field)
		content.WriteString(fmt.Sprintf("\tif !%s(r.%s) {\n", field.Validate().(string), fieldName))
		content.WriteString(fmt.Sprintf("\t\treturn fmt.Errorf(\"Validation failed for field name: %s\")\n", fieldName))
		content.WriteString("\t}\n")
	}
	content.WriteString("return nil\n")
	content.WriteString("}\n\n")
	return content.String()
}

func generateUpdateRequest(entity schema.Entity) string {
	var content strings.Builder
	content.WriteString(fmt.Sprintf("func (r *Update%sRequest) Validate() error {\n", entity.Name))
	for _, field := range entity.Fields {
		if field.Validate == nil {
			continue
		}

		fieldName := toProtoFieldName(field)
		content.WriteString(fmt.Sprintf("\tif !%s(r.%s) {\n", field.Validate().(string), fieldName))
		content.WriteString(fmt.Sprintf("\t\treturn fmt.Errorf(\"Validation failed for field name: %s\")\n", fieldName))
		content.WriteString("\t}\n")
	}
	content.WriteString("return nil\n")
	content.WriteString("}\n\n")
	return content.String()
}

func hasValidateField(entity schema.Entity) bool {
	for _, field := range entity.Fields {
		if field.Validate != nil {
			return true
		}
	}
	return false
}

func toProtoFieldName(field schema.Field) string {
	if field.IsID() {
		return strings.ToUpper(field.Name[:1]) + field.Name[1:]
	}
	return snakeToCamelCase(field.Name)
}

func snakeToCamelCase(s string) string {
	parts := strings.Split(s, "_")
	result := ""
	for _, part := range parts {
		if len(part) > 0 {
			result += strings.ToUpper(part[:1]) + part[1:]
		}
	}
	return result
}
