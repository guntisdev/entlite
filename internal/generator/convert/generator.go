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
	content.WriteString("\t\"database/sql\"\n")
	content.WriteString("\t\"google.golang.org/protobuf/types/known/timestamppb\"\n")
	for _, importPath := range imports {
		content.WriteString(fmt.Sprintf("\t%s\n", importPath))
	}
	content.WriteString(")\n\n")

	for _, entity := range messageEntities {
		content.WriteString(generateEntityConversion(entity))
		content.WriteString("\n")
	}

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

	content.WriteString(fmt.Sprintf("// %s conversion functions\n\n", entity.Name))
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

	// TODO proto to DB

	content.WriteString("}")

	return content.String()
}

func fieldDBToProto(field schema.Field, dbFieldName string, dbPrefix string) string {
	dbFieldRef := fmt.Sprintf("%s.%s", dbPrefix, dbFieldName)

	switch field.Type {
	case schema.FieldTypeString:
		return dbFieldRef
	case schema.FieldTypeInt32:
		return dbFieldRef
	case schema.FieldTypeBool:
		return dbFieldRef
	case schema.FieldTypeTime:
		return fmt.Sprintf("TimeToProtoTimestamp(&%s)", dbFieldRef)
	default:
		return dbFieldRef
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
