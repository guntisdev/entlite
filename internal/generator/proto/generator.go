package proto

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/guntisdev/entlite/internal/schema"
)

func Generate(entities []schema.Entity, dir string) error {
	var messageEntities []schema.Entity
	for _, entity := range entities {
		if entity.HasMessage() {
			messageEntities = append(messageEntities, entity)
		}
	}

	var serviceEntities []schema.Entity
	for _, entity := range entities {
		if entity.HasService() {
			serviceEntities = append(serviceEntities, entity)
		}
	}

	protoContent := generateSchemaProto(messageEntities, serviceEntities)

	fileName := "schema.proto"
	filePath := filepath.Join(dir, fileName)

	if err := writeFile(filePath, protoContent); err != nil {
		return fmt.Errorf("failed to write proto file %s: %w", fileName, err)
	}

	return nil
}

func generateSchemaProto(messageEntities []schema.Entity, serviceEntities []schema.Entity) string {
	var content strings.Builder

	content.WriteString("syntax = \"proto3\";\n\n")
	content.WriteString(fmt.Sprintf("package %s;\n\n", "entlite"))
	content.WriteString("option go_package = \".pb\";\n\n")

	imports := []string{}
	if needsCommonImports(messageEntities) {
		imports = append(imports, "google/protobuf/timestamp.proto")
	}
	if needsEmptyImportForEntities(serviceEntities) {
		imports = append(imports, "google/protobuf/empty.proto")
	}
	for _, imp := range imports {
		content.WriteString(fmt.Sprintf("import \"%s\";\n", imp))
	}
	if len(imports) > 0 {
		content.WriteString("\n")
	}

	for i, entity := range messageEntities {
		if i > 0 {
			content.WriteString("\n")
		}

		content.WriteString(fmt.Sprintf("// %s represents as %s entity\n", entity.Name, strings.ToLower(entity.Name)))
		content.WriteString(fmt.Sprintf("message %s{\n", entity.Name))

		for _, field := range entity.Fields {
			protoType := getProtoType(field.Type)
			fieldNumber := 1 // actual number should be read/generated at parser stage
			if field.ProtoField != nil {
				fieldNumber = *field.ProtoField
			}
			content.WriteString(fmt.Sprintf("  %s %s = %d;\n", protoType, field.Name, fieldNumber))
		}

		content.WriteString("}")
		if i < len(messageEntities)-1 {
			content.WriteString("\n")
		}
	}

	// Add spacing betweenn mesages and services if both exists
	if len(messageEntities) > 0 && len(serviceEntities) > 0 {
		content.WriteString("\n\n")
	}

	for i, entity := range serviceEntities {
		if i > 0 {
			content.WriteString("\n\n")
		}

		content.WriteString(generateServiceProto(entity))
	}

	return content.String()
}

func generateServiceProto(entity schema.Entity) string {
	var content strings.Builder

	return content.String()
}

func writeFile(filePath, content string) error {
	return os.WriteFile(filePath, []byte(content), 0644)
}

func needsCommonImports(entities []schema.Entity) bool {
	for _, entity := range entities {
		for _, field := range entity.Fields {
			if field.Type == schema.FieldTypeTime {
				return true
			}
		}
	}

	return false
}

func needsEmptyImportForEntities(entities []schema.Entity) bool {
	for _, entity := range entities {
		methods := entity.GetMethods()
		for _, method := range methods {
			if method == schema.MethodDelete {
				return true
			}
		}
	}

	return false
}

func getProtoType(fieldType schema.FieldType) string {
	switch fieldType {
	case schema.FieldTypeString:
		return "string"
	case schema.FieldTypeInt32:
		return "int32"
	case schema.FieldTypeBool:
		return "bool"
	case schema.FieldTypeTime:
		return "google.protobuf.Timestamp"
	default:
		return "string"
	}
}
