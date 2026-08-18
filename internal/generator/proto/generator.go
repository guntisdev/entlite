package proto

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/guntisdev/entlite/internal/schema"
	"github.com/guntisdev/entlite/internal/util"
	"github.com/guntisdev/entlite/pkg/entlite/permissions"
)

func Generate(entities []schema.Entity, dir string) error {
	protoContent := generateSchemaProto(entities)

	fileName := "schema.proto"
	filePath := filepath.Join(dir, fileName)

	if err := writeFile(filePath, protoContent); err != nil {
		return fmt.Errorf("failed to write proto file %s: %w", fileName, err)
	}

	return nil
}

func generateSchemaProto(entities []schema.Entity) string {
	var content strings.Builder

	content.WriteString("syntax = \"proto3\";\n\n")
	content.WriteString(fmt.Sprintf("package %s;\n\n", "entlite"))
	content.WriteString("option go_package = \"./pb\";\n\n")

	imports := []string{}
	if needsCommonImports(entities) {
		imports = append(imports, "google/protobuf/timestamp.proto")
	}
	if needsEmptyImportForEntities(entities) {
		imports = append(imports, "google/protobuf/empty.proto")
	}
	imports = append(imports, "buf/validate/validate.proto")
	for _, imp := range imports {
		content.WriteString(fmt.Sprintf("import \"%s\";\n", imp))
	}
	if len(imports) > 0 {
		content.WriteString("\n")
	}

	var requiredStr = "[(buf.validate.field).required = true]"

	for i, entity := range entities {
		if i > 0 {
			content.WriteString("\n")
		}

		content.WriteString(fmt.Sprintf("// %s represents as %s entity\n", entity.Name, strings.ToLower(entity.Name)))
		content.WriteString(fmt.Sprintf("message %s {\n", entity.Name))

		for _, field := range entity.Fields {
			canRead := (field.Permissions & permissions.ApiRead) != 0
			if !canRead {
				continue
			}
			writeFieldComment(&content, field.Comment)
			protoType := getProtoType(field.Type)
			var optional string
			var required string
			if field.Optional {
				optional = "optional "
			} else if field.Type == schema.FieldTypeBool {
				// proto does not differentiat between bool undefined or false
				required = ""
			} else {
				required = fmt.Sprintf(" %s", requiredStr)
			}
			content.WriteString(fmt.Sprintf("  %s%s %s = %d%s;\n", optional, protoType, field.Name, field.ProtoField, required))
		}

		content.WriteString("}")
		if i < len(entities)-1 {
			content.WriteString("\n")
		}
	}

	// Add spacing between messages and services
	if len(entities) > 0 {
		content.WriteString("\n\n")
	}

	for i, entity := range entities {
		if i > 0 {
			content.WriteString("\n\n")
		}

		content.WriteString(generateServiceProto(entity))
	}

	return content.String()
}

func generateServiceProto(entity schema.Entity) string {
	var content strings.Builder

	content.WriteString(generateResponseMessages(entity))
	content.WriteString("\n\n")

	serviceName := fmt.Sprintf("%sService", entity.Name)
	content.WriteString(fmt.Sprintf("// %s provides CRUD opertions for %s entities\n", serviceName, entity.Name))
	content.WriteString(fmt.Sprintf("service %s {\n", serviceName))

	for _, query := range entity.Queries {
		content.WriteString(generateRequests(entity, query))
	}

	content.WriteString("}")

	return content.String()
}

func generateResponseMessages(entity schema.Entity) string {
	var content strings.Builder
	var requiredStr = "[(buf.validate.field).required = true]"

	for i, query := range entity.Queries {
		if i > 0 {
			content.WriteString("\n")
		}

		messageName := util.GenQueryName(query, entity.Name)

		switch query.Type {
		case schema.QueryCreate:
			content.WriteString(fmt.Sprintf("message %sRequest {\n", messageName))
			writeCreateFields(&content, entity)
			content.WriteString("}")
		case schema.QueryCreateBulk:
			content.WriteString(fmt.Sprintf("message %sItem {\n", messageName))
			writeCreateFields(&content, entity)
			content.WriteString("}\n\n")

			content.WriteString(fmt.Sprintf("message %sRequest {\n", messageName))
			content.WriteString(fmt.Sprintf("  repeated %sItem items = 1 %s;\n", messageName, requiredStr))
			content.WriteString("}\n\n")

			content.WriteString(fmt.Sprintf("message %sResponse {\n", messageName))
			content.WriteString(fmt.Sprintf("  repeated %s %ss = 1;\n", entity.Name, strings.ToLower(entity.Name)))
			content.WriteString("}")
		case schema.QueryGetBy:
			content.WriteString(fmt.Sprintf("message %sRequest {\n", messageName))

			for _, fieldName := range query.Fields {
				field, found := entity.GetFieldByName(fieldName)
				if !found {
					continue
				}

				protoType := getProtoType(field.Type)
				content.WriteString(fmt.Sprintf("  %s %s = %d %s;\n", protoType, field.Name, field.ProtoField, requiredStr))
			}
			content.WriteString("}")
		case schema.QueryUpdate:
			content.WriteString(fmt.Sprintf("message %sRequest {\n", messageName))
			for _, field := range entity.Fields {
				canWrite := (field.Permissions & permissions.ApiWrite) != 0
				if !field.IsID() {
					if field.Immutable || !canWrite {
						continue
					}
				}
				writeFieldComment(&content, field.Comment)
				protoType := getProtoType(field.Type)
				var optional string
				var required string
				// special case for psw etc - if not readable then no obligatory to update
				canRead := (field.Permissions & permissions.ApiRead) != 0
				if field.Optional || !canRead || field.DefaultValue != nil || field.DefaultFunc != nil {
					optional = "optional "
				} else {
					required = fmt.Sprintf(" %s", requiredStr)
				}
				content.WriteString(fmt.Sprintf("  %s%s %s = %d%s;\n", optional, protoType, field.Name, field.ProtoField, required))
			}
			content.WriteString("}")
		case schema.QueryDelete:
			content.WriteString(fmt.Sprintf("message %sRequest {\n", messageName))
			content.WriteString(fmt.Sprintf("  %s %s;\n", getIdFieldAsStr(entity.Fields), requiredStr))
			content.WriteString("}")
		case schema.QueryDeleteAll:
			content.WriteString(fmt.Sprintf("message %sRequest {\n", messageName))
			content.WriteString("}")
		case schema.QueryListAll:
			content.WriteString(fmt.Sprintf("message %sRequest {\n", messageName))
			content.WriteString("}\n\n")

			content.WriteString(fmt.Sprintf("message %sResponse {\n", messageName))
			content.WriteString(fmt.Sprintf("  repeated %s %ss = 1;\n", entity.Name, strings.ToLower(entity.Name)))
			content.WriteString("}")
		case schema.QueryListBy:
			content.WriteString(fmt.Sprintf("message %sRequest {\n", messageName))
			// TODO proly change int type depending on ID field type
			content.WriteString(fmt.Sprintf("  int32 limit = 1 %s;\n", requiredStr))
			content.WriteString("  int32 offset = 2;\n")

			protoFieldNum := 3
			for _, fieldName := range query.Fields {
				field, found := entity.GetFieldByName(fieldName)
				if !found {
					continue
				}

				protoType := getProtoType(field.Type)
				content.WriteString(fmt.Sprintf("  %s %s = %d %s;\n", protoType, field.Name, protoFieldNum, requiredStr))
				protoFieldNum++
			}
			for _, filter := range query.Filters {
				field, found := entity.GetFieldByName(filter.Field)
				if !found {
					continue
				}

				protoType := getProtoType(field.Type)

				// Range filters expand to min_/max_ params, matching sqlc.
				var names []string
				if filter.Type == schema.QueryFilterRange {
					names = []string{"min_" + filter.Field, "max_" + filter.Field}
				} else {
					names = []string{filter.Field}
				}

				for _, name := range names {
					if filter.Optional {
						content.WriteString(fmt.Sprintf("  optional %s %s = %d;\n", protoType, name, protoFieldNum))
					} else {
						content.WriteString(fmt.Sprintf("  %s %s = %d %s;\n", protoType, name, protoFieldNum, requiredStr))
					}
					protoFieldNum++
				}
			}
			content.WriteString("}\n\n")

			content.WriteString(fmt.Sprintf("message %sResponse {\n", messageName))
			content.WriteString(fmt.Sprintf("  repeated %s %ss = 1;\n", entity.Name, strings.ToLower(entity.Name)))
			content.WriteString("}")
		}

	}

	return content.String()
}

func writeFieldComment(content *strings.Builder, comment string) {
	if comment == "" {
		return
	}
	for line := range strings.SplitSeq(comment, "\n") {
		fmt.Fprintf(content, "  // %s\n", strings.TrimRight(line, "\r"))
	}
}

func writeCreateFields(content *strings.Builder, entity schema.Entity) {
	var requiredStr = "[(buf.validate.field).required = true]"
	for _, field := range entity.Fields {
		canWrite := (field.Permissions & permissions.ApiWrite) != 0
		if field.IsID() || !canWrite {
			continue
		}
		writeFieldComment(content, field.Comment)
		protoType := getProtoType(field.Type)
		var optional string
		var required string
		if field.Optional || field.DefaultValue != nil || field.DefaultFunc != nil {
			optional = "optional "
		} else if field.Type == schema.FieldTypeBool {
			required = ""
		} else {
			required = fmt.Sprintf(" %s", requiredStr)
		}
		content.WriteString(fmt.Sprintf("  %s%s %s = %d%s;\n", optional, protoType, field.Name, field.ProtoField, required))
	}
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

func generateRequests(entity schema.Entity, query schema.Query) string {
	rpcName := util.GenQueryRpcName(query, entity.Name)
	messageName := util.GenQueryName(query, entity.Name)

	switch query.Type {
	case schema.QueryCreate:
		return fmt.Sprintf("  rpc %s(%sRequest) returns (%s);\n", rpcName, messageName, entity.Name)
	case schema.QueryCreateBulk:
		return fmt.Sprintf("  rpc %s(%sRequest) returns (%sResponse);\n", rpcName, messageName, messageName)
	case schema.QueryGetBy:
		return fmt.Sprintf("  rpc %s(%sRequest) returns (%s);\n", rpcName, messageName, entity.Name)
	case schema.QueryUpdate:
		return fmt.Sprintf("  rpc %s(%sRequest) returns (%s);\n", rpcName, messageName, entity.Name)
	case schema.QueryDelete, schema.QueryDeleteAll:
		return fmt.Sprintf("  rpc %s(%sRequest) returns (google.protobuf.Empty);\n", rpcName, messageName)
	case schema.QueryListBy, schema.QueryListAll:
		return fmt.Sprintf("  rpc %s(%sRequest) returns (%sResponse);\n", rpcName, messageName, messageName)
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
		for _, query := range entity.Queries {
			if query.Type == schema.QueryDelete || query.Type == schema.QueryDeleteAll {
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
	case schema.FieldTypeInt:
		return "int32"
	case schema.FieldTypeInt64:
		return "int64"
	case schema.FieldTypeFloat:
		return "double"
	case schema.FieldTypeBool:
		return "bool"
	case schema.FieldTypeTime:
		return "google.protobuf.Timestamp"
	case schema.FieldTypeByte:
		return "bytes"
	case schema.FieldTypeJSON:
		return "string" // raw json text, kept as string so it passes through unchanged
	default:
		return "string"
	}
}
