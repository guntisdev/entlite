package sqlcwrap

import (
	"fmt"
	"go/ast"
	"strconv"
	"strings"

	"github.com/guntisdev/entlite/internal/generator/sqlc"
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
		zeroValue = "nil"
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

func sqlToGo(field schema.Field, pbFieldRef string, sqlDialect sqlc.SQLDialect) string {
	if sqlDialect == sqlc.SQLite {
		if field.Type == schema.FieldTypeBool {
			return fmt.Sprintf("SQLiteBoolToInt(%s)", pbFieldRef)
		}
		if field.Type == schema.FieldTypeInt {
			if field.Optional {
				return fmt.Sprintf("IntPtrConvert[%s, %s](%s)", "int32", "int64", pbFieldRef)
			} else {
				return fmt.Sprintf("IntConvert[%s, %s](%s)", "int32", "int64", pbFieldRef)
			}
		}
	}

	if sqlDialect == sqlc.MySQL && field.Optional {
		if field.Type == schema.FieldTypeByte {
			// Special case: some generated refs may be dereferenced (e.g. *arg.ApiKey),
			// but PtrBytesToNullString expects a pointer, so strip a leading '*'.
			normalizedRef := strings.TrimPrefix(pbFieldRef, "*")
			return fmt.Sprintf("PtrBytesToNullString(%s)", normalizedRef)
		}
	}

	if field.Optional && (sqlDialect == sqlc.PostgreSQL || sqlDialect == sqlc.MySQL) {
		switch field.Type {
		case schema.FieldTypeString:
			return fmt.Sprintf("PtrToNullString(%s)", pbFieldRef)
		case schema.FieldTypeInt:
			return fmt.Sprintf("PtrToNullInt32(%s)", pbFieldRef)
		case schema.FieldTypeInt64:
			return fmt.Sprintf("PtrToNullInt64(%s)", pbFieldRef)
		case schema.FieldTypeFloat:
			return fmt.Sprintf("PtrToNullFloat64(%s)", pbFieldRef)
		case schema.FieldTypeBool:
			return fmt.Sprintf("PtrToNullBool(%s)", pbFieldRef)
		}
	}

	return pbFieldRef
}

// goFromSQL converts from SQL types to Go types (inverse of sqlToGo)
func goFromSQL(field schema.Field, dbFieldRef string, sqlDialect sqlc.SQLDialect) string {
	if sqlDialect == sqlc.SQLite {
		if field.Type == schema.FieldTypeBool {
			return fmt.Sprintf("SQLiteIntToBool(%s)", dbFieldRef)
		}
		if field.Type == schema.FieldTypeInt {
			if field.Optional {
				return fmt.Sprintf("IntPtrConvert[%s, %s](%s)", "int64", "int32", dbFieldRef)
			} else {
				return fmt.Sprintf("IntConvert[%s, %s](%s)", "int64", "int32", dbFieldRef)
			}
		}
	}

	if field.Optional && (sqlDialect == sqlc.PostgreSQL || sqlDialect == sqlc.MySQL) {
		switch field.Type {
		case schema.FieldTypeString:
			return fmt.Sprintf("NullStringToPtr(%s)", dbFieldRef)
		case schema.FieldTypeInt:
			return fmt.Sprintf("NullInt32ToPtr(%s)", dbFieldRef)
		case schema.FieldTypeInt64:
			return fmt.Sprintf("NullInt64ToPtr(%s)", dbFieldRef)
		case schema.FieldTypeFloat:
			return fmt.Sprintf("NullFloat64ToPtr(%s)", dbFieldRef)
		case schema.FieldTypeBool:
			return fmt.Sprintf("NullBoolToPtr(%s)", dbFieldRef)
		}
	}

	return dbFieldRef
}

func formatDefaultValue(field schema.Field) string {
	switch v := field.DefaultValue.(type) {
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case float32:
		return strconv.FormatFloat(float64(v), 'f', -1, 32)
	case int32:
		return fmt.Sprintf("%d", v)
	case int64:
		return fmt.Sprintf("%d", v)
	case int:
		return fmt.Sprintf("%d", v)
	case bool:
		return fmt.Sprintf("%v", v)
	case string:
		return fmt.Sprintf("%q", v)
	default:
		return fmt.Sprintf("%v", v)
	}
}
