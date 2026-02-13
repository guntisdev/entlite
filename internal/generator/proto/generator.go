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
		content.WriteString(fmt.Sprintf("message %s {\n", entity.Name))

		for _, field := range entity.Fields {
			protoType := getProtoType(field.Type)
			if field.Comment != "" {
				content.WriteString(fmt.Sprintf("  // %s\n", field.Comment))
			}
			content.WriteString(fmt.Sprintf("  %s %s = %d;\n", protoType, field.Name, field.ProtoField))
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

	content.WriteString(generateServiceMessages(entity))
	content.WriteString("\n\n")

	serviceName := fmt.Sprintf("%sService", entity.Name)
	content.WriteString(fmt.Sprintf("// %s provides CRUD opertions for %s entities\n", serviceName, entity.Name))
	content.WriteString(fmt.Sprintf("service %s {\n", serviceName))

	for _, method := range entity.GetMethods() {
		content.WriteString(generateServiceMethod(entity.Name, method))
	}

	content.WriteString("}")

	return content.String()
}

func generateServiceMessages(entity schema.Entity) string {
	var content strings.Builder

	for i, method := range entity.GetMethods() {
		if i > 0 {
			content.WriteString("\n")
		}

		switch method {
		case schema.MethodCreate:
			content.WriteString(fmt.Sprintf("message Create%sRequest {\n", entity.Name))
			for _, field := range entity.Fields {
				if field.IsID() {
					continue
				}
				protoType := getProtoType(field.Type)
				if field.Comment != "" {
					content.WriteString(fmt.Sprintf("  // %s\n", field.Comment))
				}
				content.WriteString(fmt.Sprintf("  %s %s = %d;\n", protoType, field.Name, field.ProtoField))
			}
			content.WriteString("}")
		case schema.MethodGet:
			content.WriteString(fmt.Sprintf("message Get%sRequest {\n", entity.Name))
			content.WriteString(fmt.Sprintf("  %s;\n", getIdFieldAsStr(entity.Fields)))
			content.WriteString("}")
		case schema.MethodUpdate:
			content.WriteString(fmt.Sprintf("message Update%sRequest {\n", entity.Name))
			for _, field := range entity.Fields {
				if field.Immutable && !field.IsID() {
					continue
				}
				protoType := getProtoType(field.Type)
				if field.Comment != "" {
					content.WriteString(fmt.Sprintf("  // %s\n", field.Comment))
				}
				content.WriteString(fmt.Sprintf("  %s %s = %d;\n", protoType, field.Name, field.ProtoField))
			}
			content.WriteString("}")
		case schema.MethodDelete:
			content.WriteString(fmt.Sprintf("message Delete%sRequest {\n", entity.Name))
			content.WriteString(fmt.Sprintf("  %s;\n", getIdFieldAsStr(entity.Fields)))
			content.WriteString("}")
		case schema.MethodList:
			content.WriteString(fmt.Sprintf("message List%sRequest {\n", entity.Name))
			content.WriteString("  int32 limit = 1;\n")
			content.WriteString("  int32 offset = 2;\n")
			content.WriteString("}\n\n")

			content.WriteString(fmt.Sprintf("message List%sResponse {\n", entity.Name))
			content.WriteString(fmt.Sprintf("  repeated %s %ss = 1;\n", entity.Name, strings.ToLower(entity.Name)))
			content.WriteString("}")
		}

	}

	return content.String()
}

func getIdFieldAsStr(fields []schema.Field) string {
	for _, field := range fields {
		if field.IsID() {
			protoType := getProtoType(field.Type)
			return fmt.Sprintf("%s %s = %d", protoType, field.Name, field.ProtoField)
		}
	}

	return "int32 id = 1"
}

func generateServiceMethod(entityName string, method schema.Method) string {
	switch method {
	case schema.MethodCreate:
		return fmt.Sprintf("  rpc Create(Create%sRequest) returns (%s);\n", entityName, entityName)
	case schema.MethodGet:
		return fmt.Sprintf("  rpc Get(Get%sRequest) returns (%s);\n", entityName, entityName)
	case schema.MethodUpdate:
		return fmt.Sprintf("  rpc Update(Update%sRequest) returns (%s);\n", entityName, entityName)
	case schema.MethodDelete:
		return fmt.Sprintf("  rpc Delete(Delete%sRequest) returns (google.protobuf.Empty);\n", entityName)
	case schema.MethodList:
		return fmt.Sprintf("  rpc List(List%sRequest) returns (List%sUser);\n", entityName, entityName)
	default:
		return ""
	}
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
