package convert

import (
	"fmt"
	"strings"

	"github.com/guntisdev/entlite/internal/schema"
)

func Generate(entities []schema.Entity, imports []string) (string, error) {
	var messageEntities []schema.Entity
	for _, entity := range entities {
		if entity.HasMessage() {
			messageEntities = append(messageEntities, entity)
		}
	}

	if len(messageEntities) == 0 {
		return "", fmt.Errorf("No entities with Message annotation found for convert.go\n")
	}

	var content strings.Builder

	content.WriteString("// generate convertion between db and pb types\n")
	content.WriteString("package convert\n\n")

	content.WriteString("import (\n")
	// TODO find this import dynamically
	content.WriteString("\t\"database/sql\"\n")

	// TODO check from DefaultFunc actual imports
	content.WriteString("\t\"time\"\n")
	content.WriteString("\t\"google.golang.org/protobuf/types/known/timestamppb\"\n")
	for _, importPath := range imports {
		content.WriteString(fmt.Sprintf("\t%s\n", importPath))
	}
	content.WriteString(")\n\n")

	for _, entity := range messageEntities {
		content.WriteString(generateEntityConversion(entity))
		content.WriteString("\n")
	}

	content.WriteString("\n\n")
	content.WriteString("// ++++++ Helper functions for type conversions\n")
	content.WriteString(generateHelperFunctions())

	return content.String(), nil
}

func generateEntityConversion(entity schema.Entity) string {
	var content strings.Builder

	// TODO maybe extract prefix from imports strings someday?
	dbPrefix := "db"
	pbPrefix := "pb"
	dbName := fmt.Sprintf("%s.%s", dbPrefix, entity.Name)
	pbName := fmt.Sprintf("%s.%s", pbPrefix, entity.Name)

	content.WriteString(fmt.Sprintf("// +++++ %s conversion functions\n\n", entity.Name))
	content.WriteString(fmt.Sprintf("// %s DBToProto converts a database model to proto message\n", entity.Name))
	content.WriteString(fmt.Sprintf("func %sDBToProto(db *%s) *%s {\n", entity.Name, dbName, pbName))
	content.WriteString("\tif db == nil {\n")
	content.WriteString("\t\treturn nil\n")
	content.WriteString("\t}\n\n")
	content.WriteString(fmt.Sprintf("\treturn &%s{\n", pbName))

	for _, field := range entity.Fields {
		dbFieldName := toDBFieldName(field)
		conversion := fieldDBToProto(field, dbFieldName, dbPrefix)
		protoFieldName := toProtoFieldName(field)
		content.WriteString(fmt.Sprintf("\t\t%s: %s,\n", protoFieldName, conversion))
	}

	content.WriteString("\t}\n")
	content.WriteString("}\n\n")

	content.WriteString(fmt.Sprintf("// %sProtoToDB converts a proto message to database model\n", entity.Name))
	content.WriteString(fmt.Sprintf("func %sProtoToDB(pb *%s) *%s {\n", entity.Name, pbName, dbName))
	content.WriteString("\tif pb == nil {\n")
	content.WriteString("\t\treturn nil\n")
	content.WriteString("\t}\n\n")
	content.WriteString(fmt.Sprintf("\treturn &%s{\n", dbName))

	for _, field := range entity.Fields {
		protoFieldName := toProtoFieldName(field)
		conversion := fieldProtoToDB(field, protoFieldName, pbPrefix)
		dbFieldName := toDBFieldName(field)
		content.WriteString(fmt.Sprintf("\t\t%s: %s,\n", dbFieldName, conversion))
	}

	content.WriteString("\t}\n")
	content.WriteString("}\n")

	return content.String()
}

func fieldDBToProto(field schema.Field, dbFieldName string, dbPrefix string) string {
	dbFieldRef := fmt.Sprintf("%s.%s", dbPrefix, dbFieldName)

	if field.Optional {
		switch field.Type {
		case schema.FieldTypeString:
			return fmt.Sprintf("NullStringToPtr(%s)", dbFieldRef)
		case schema.FieldTypeInt32:
			return fmt.Sprintf("NullInt32ToPtr(%s)", dbFieldRef)
		case schema.FieldTypeBool:
			return fmt.Sprintf("NullBoolToPtr(%s)", dbFieldRef)
		case schema.FieldTypeTime:
			return fmt.Sprintf("NullTimeToProto(%s)", dbFieldRef)
		default:
			return dbFieldRef
		}
	}

	switch field.Type {
	case schema.FieldTypeString:
		return dbFieldRef
	case schema.FieldTypeInt32:
		return dbFieldRef
	case schema.FieldTypeBool:
		return dbFieldRef
	case schema.FieldTypeTime:
		return fmt.Sprintf("TimeToProto(%s)", dbFieldRef)
	default:
		return dbFieldRef
	}
}

func fieldProtoToDB(field schema.Field, protoFieldName string, pbPrefix string) string {
	pbFieldRef := fmt.Sprintf("%s.%s", pbPrefix, protoFieldName)

	if field.Optional {
		switch field.Type {
		case schema.FieldTypeString:
			return fmt.Sprintf("PtrToNullString(%s)", pbFieldRef)
		case schema.FieldTypeInt32:
			return fmt.Sprintf("PtrToNullInt32(%s)", pbFieldRef)
		case schema.FieldTypeBool:
			return fmt.Sprintf("PtrToNullBool(%s)", pbFieldRef)
		case schema.FieldTypeTime:
			return fmt.Sprintf("ProtoToNullTime(%s)", pbFieldRef)
		default:
			return pbFieldRef
		}
	}

	switch field.Type {
	case schema.FieldTypeString:
		return pbFieldRef
	case schema.FieldTypeInt32:
		return pbFieldRef
	case schema.FieldTypeBool:
		return pbFieldRef
	case schema.FieldTypeTime:
		return fmt.Sprintf("ProtoToTime(%s)", pbFieldRef)
	default:
		return pbFieldRef
	}
}

func toProtoFieldName(field schema.Field) string {
	if field.IsID() {
		return strings.ToUpper(field.Name[:1]) + field.Name[1:]
	}
	return snakeToCamelCase(field.Name)
}

// match sqlc conversion - ID and CamelCase names
func toDBFieldName(field schema.Field) string {
	if field.IsID() {
		return "ID"
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
