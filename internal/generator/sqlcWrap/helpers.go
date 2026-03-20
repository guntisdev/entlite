package sqlcwrap

import (
	"fmt"
	"go/ast"
	"strings"

	"github.com/guntisdev/entlite/internal/schema"
)

func fieldToGoType(field schema.Field) string {
	optionalStr := ""
	if field.Optional {
		optionalStr = "*"
	}

	switch field.Type {
	case schema.FieldTypeString:
		return fmt.Sprintf("%sstring", optionalStr)
	case schema.FieldTypeInt:
		return fmt.Sprintf("%sint32", optionalStr)
	case schema.FieldTypeInt64:
		return fmt.Sprintf("%sint64", optionalStr)
	case schema.FieldTypeFloat:
		return fmt.Sprintf("%sfloat64", optionalStr)
	case schema.FieldTypeBool:
		return fmt.Sprintf("%sbool", optionalStr)
	case schema.FieldTypeTime:
		return fmt.Sprintf("%stime.Time", optionalStr)
	case schema.FieldTypeByte:
		return fmt.Sprintf("%s[]byte", optionalStr)
	default:
		return fmt.Sprintf("%sstring", optionalStr)
	}
}

func getFieldByName(entity schema.Entity, name string) *schema.Field {
	for _, field := range entity.Fields {
		fmt.Printf(">> %s %s\n", field.Name, name)
		if toDBFieldName(field) == name {
			return &field
		}
	}
	return nil
}

func addValidationChecks(entity schema.Entity, sqlQuery string, returnType string) string {
	var sb strings.Builder

	var zeroValue string
	switch returnType {
	case "", "int", "int8", "int16", "int32", "int64", "uint", "uint8", "uint16", "uint32", "uint64":
		zeroValue = "0"
	case "float32", "float64":
		zeroValue = "0.0"
	case "bool":
		zeroValue = "false"
	case "string":
		zeroValue = "\"\""
	default:
		// Handle pointers, slices, and maps
		if strings.HasPrefix(returnType, "*") || strings.HasPrefix(returnType, "[]") || strings.HasPrefix(returnType, "map[") {
			zeroValue = "nil"
		} else {
			// Assume it's a struct type
			zeroValue = returnType + "{}"
		}
	}

	for _, field := range entity.Fields {
		if field.Validate == nil {
			continue
		}

		validateName := field.Validate().(string)
		fieldName := toDBFieldName(field)
		sb.WriteString(fmt.Sprintf("\tif !%s(arg.%s) {\n", validateName, fieldName))
		sb.WriteString(fmt.Sprintf("\t\treturn %s, fmt.Errorf(\"Failed %s: incorrect value for '%s' in field '%s', validated by '%s'\")\n", zeroValue, sqlQuery, entity.Name, field.Name, validateName))
		sb.WriteString("\t}\n")
	}
	return sb.String()
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

func hasValidateField(entity schema.Entity) bool {
	for _, field := range entity.Fields {
		if field.Validate != nil {
			return true
		}
	}
	return false
}

func hasDefaultFuncFields(entity schema.Entity) bool {
	for _, field := range entity.Fields {
		if field.DefaultFunc != nil {
			return true
		}
	}
	return false
}

// for example updated_at field
func hasDefaultFuncAndNoImmutable(entity schema.Entity) bool {
	for _, field := range entity.Fields {
		if field.DefaultFunc != nil && !field.Immutable {
			return true
		}
	}
	return false
}

func usesSQLTypes(structType *ast.StructType) bool {
	for _, field := range structType.Fields.List {
		if usesSQLType(field.Type) {
			return true
		}
	}
	return false
}

func usesSQLType(expr ast.Expr) bool {
	switch t := expr.(type) {
	case *ast.SelectorExpr:
		if ident, ok := t.X.(*ast.Ident); ok {
			return ident.Name == "sql"
		}
	case *ast.StarExpr:
		return usesSQLType(t.X)
	case *ast.ArrayType:
		return usesSQLType(t.Elt)
	}
	return false
}

func formatType(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return "*" + formatType(t.X)
	case *ast.SelectorExpr:
		return formatType(t.X) + "." + t.Sel.Name
	case *ast.ArrayType:
		return "[]" + formatType(t.Elt)
	default:
		return "interface{}"
	}
}

func toExportedName(name string) string {
	parts := strings.Split(name, "_")
	for i := range parts {
		if len(parts[i]) > 0 {
			parts[i] = strings.ToUpper(parts[i][:1]) + parts[i][1:]
		}
	}
	return strings.Join(parts, "")
}
