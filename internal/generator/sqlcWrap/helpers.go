package sqlcwrap

import (
	"fmt"
	"go/ast"
	"strconv"
	"strings"

	"github.com/guntisdev/entlite/internal/schema"
)

func fieldToGoType(field schema.Field) string {
	optionalStr := ""
	if field.Optional {
		optionalStr = "*"
	}

	switch field.Type {
	case schema.FieldTypeString, schema.FieldTypeJSON:
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

// converts query sql types to go type
func filterParamField(entity schema.Entity, paramName string) (schema.Field, bool) {
	lookup := func(name string) (schema.Field, bool) {
		for _, field := range entity.Fields {
			if strings.EqualFold(toDBFieldName(field), name) {
				return field, true
			}
		}
		return schema.Field{}, false
	}

	if field, ok := lookup(paramName); ok {
		return field, true
	}

	for _, prefix := range []string{"Min", "Max"} {
		if len(paramName) > len(prefix) && strings.EqualFold(paramName[:len(prefix)], prefix) {
			if field, ok := lookup(paramName[len(prefix):]); ok {
				return field, true
			}
		}
	}

	return schema.Field{}, false
}

// restates sqlc "<Query>Params" struct in wrapper's own types, keeping sqlc's field names and json tags.
func generateFilterParamsStruct(structName string, structType *ast.StructType, entity schema.Entity) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("type %s struct {\n", structName))

	for _, astField := range structType.Fields.List {
		if len(astField.Names) == 0 {
			continue
		}
		fieldName := astField.Names[0].Name

		goType := formatType(astField.Type)
		if field, ok := filterParamField(entity, fieldName); ok {
			goType = fieldToGoType(field)
		}

		sb.WriteString(fmt.Sprintf("\t%s %s", fieldName, goType))
		if astField.Tag != nil {
			sb.WriteString(fmt.Sprintf(" %s", astField.Tag.Value))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("}\n\n")
	return sb.String()
}

// builds internal params literal handed to sqlc, converting each field back to its dialect type.
func generateFilterParamsArg(structName string, structType *ast.StructType, entity schema.Entity, inputPkg, argVar string, sqlDialect schema.SQLDialect) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("\tinternalArg := %s.%s{\n", inputPkg, structName))

	for _, astField := range structType.Fields.List {
		if len(astField.Names) == 0 {
			continue
		}
		fieldName := astField.Names[0].Name

		valueRef := fmt.Sprintf("%s.%s", argVar, fieldName)
		if field, ok := filterParamField(entity, fieldName); ok {
			valueRef = sqlToGo(field, valueRef, sqlDialect)
		}

		sb.WriteString(fmt.Sprintf("\t\t%s: %s,\n", fieldName, valueRef))
	}

	sb.WriteString("\t}\n")
	return sb.String()
}

// renders a get/list wrapper's signature params, the arguments
// forwarded to the sqlc method, and any statements needed before the call.
func (ctx *generationContext) wrapFilterParams(funcDecl *ast.FuncDecl, entity schema.Entity) (params, args, prelude string) {
	if funcDecl.Type.Params == nil {
		return "", "", ""
	}

	var paramsSb, argsSb, preludeSb strings.Builder

	// Index 0 is ctx, which callers emit themselves.
	for i := 1; i < len(funcDecl.Type.Params.List); i++ {
		param := funcDecl.Type.Params.List[i]
		typeName := formatType(param.Type)

		for _, name := range param.Names {
			// A params struct the wrapper restates: take ours, convert to sqlc's.
			if structType, ok := ctx.filterParamsStructs[typeName]; ok {
				paramsSb.WriteString(fmt.Sprintf(", %s %s", name.Name, typeName))
				preludeSb.WriteString(generateFilterParamsArg(typeName, structType, entity, ctx.inputPackageName, name.Name, ctx.sqlDialect))
				argsSb.WriteString(", internalArg")
				continue
			}

			// A lone filter arrives as a bare scalar rather than a struct.
			if field, ok := filterParamField(entity, name.Name); ok {
				paramsSb.WriteString(fmt.Sprintf(", %s %s", name.Name, fieldToGoType(field)))
				argsSb.WriteString(fmt.Sprintf(", %s", sqlToGo(field, name.Name, ctx.sqlDialect)))
				continue
			}

			paramsSb.WriteString(fmt.Sprintf(", %s %s", name.Name, typeName))
			argsSb.WriteString(fmt.Sprintf(", %s", name.Name))
		}
	}

	return paramsSb.String(), argsSb.String(), preludeSb.String()
}

func addValidationChecks(entity schema.Entity, sqlQuery string, returnType, argVar, indent string) string {
	return addValidationChecksIndexed(entity, sqlQuery, returnType, argVar, indent, "")
}

func addValidationChecksIndexed(entity schema.Entity, sqlQuery string, returnType, argVar, indent, indexVar string) string {
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

	itemPrefix, itemArgs := "", ""
	if indexVar != "" {
		itemPrefix = "item %d: "
		itemArgs = ", " + indexVar
	}

	// json text is checked before it reaches the db
	for _, field := range entity.Fields {
		if field.Type != schema.FieldTypeJSON || field.IsVirtual() {
			continue
		}
		if !field.CanApiWrite() {
			continue
		}
		// update skips immutable fields, so they are not in the params struct
		if sqlQuery == "update" && field.Immutable {
			continue
		}

		ref := fmt.Sprintf("%s.%s", argVar, toDBFieldName(field))
		cond := fmt.Sprintf("!json.Valid([]byte(%s))", ref)
		if isPointerParam(field, sqlQuery) {
			cond = fmt.Sprintf("%s != nil && !json.Valid([]byte(*%s))", ref, ref)
		}
		sb.WriteString(fmt.Sprintf("%sif %s {\n", indent, cond))
		sb.WriteString(fmt.Sprintf("%s\treturn %s, fmt.Errorf(\"Failed %s: %sinvalid json for '%s' in field '%s'\"%s)\n", indent, zeroValue, sqlQuery, itemPrefix, entity.Name, field.Name, itemArgs))
		sb.WriteString(fmt.Sprintf("%s}\n", indent))
	}

	// TODO fix Optional() with Validate() - a pointer is passed to a value func and does not compile
	for _, field := range entity.Fields {
		if field.Validate == nil {
			continue
		}
		if field.IsVirtual() {
			continue
		}

		validateName := field.Validate().(string)
		fieldName := toDBFieldName(field)
		sb.WriteString(fmt.Sprintf("%sif !%s(%s.%s) {\n", indent, validateName, argVar, fieldName))
		sb.WriteString(fmt.Sprintf("%s\treturn %s, fmt.Errorf(\"Failed %s: %sincorrect value for '%s' in field '%s', validated by '%s'\"%s)\n", indent, zeroValue, sqlQuery, itemPrefix, entity.Name, field.Name, validateName, itemArgs))
		sb.WriteString(fmt.Sprintf("%s}\n", indent))
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
		if len(part) == 0 {
			continue
		}
		// sqlc converts to capital ID
		if strings.ToLower(part) == "id" {
			result += "ID"
			continue
		}
		result += strings.ToUpper(part[:1]) + part[1:]
	}
	return result
}

// toProtoFieldName matches protoc-gen-go's Go struct field naming, which does
// NOT apply Go initialisms (e.g. sensor_id -> SensorId, not SensorID).
func toProtoFieldName(field schema.Field) string {
	return protoGoCamelCase(field.Name)
}

// protoGoCamelCase is a copy of google.golang.org/protobuf/internal/strs.GoCamelCase,
func protoGoCamelCase(s string) string {
	var b []byte
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case c == '.' && i+1 < len(s) && isASCIILower(s[i+1]):
			// Skip over '.' in ".{{lowercase}}".
		case c == '.':
			b = append(b, '_') // convert '.' to '_'
		case c == '_' && (i == 0 || s[i-1] == '.'):
			// Convert initial '_' to ensure we start with a capital letter.
			b = append(b, 'X') // convert '_' to 'X'
		case c == '_' && i+1 < len(s) && isASCIILower(s[i+1]):
			// Skip over '_' in "_{{lowercase}}".
		case isASCIIDigit(c):
			b = append(b, c)
		default:
			// Assume we have a letter now - if not, it's a bogus identifier.
			// The next word is a sequence of characters that must start upper case.
			if isASCIILower(c) {
				c -= 'a' - 'A' // convert lowercase to uppercase
			}
			b = append(b, c)

			// Accept lower case sequence that follows.
			for ; i+1 < len(s) && isASCIILower(s[i+1]); i++ {
				b = append(b, s[i+1])
			}
		}
	}
	return string(b)
}

func isASCIILower(c byte) bool { return 'a' <= c && c <= 'z' }
func isASCIIDigit(c byte) bool { return '0' <= c && c <= '9' }

// params are pointers when the field is optional or gets a default
func isPointerParam(field schema.Field, sqlQuery string) bool {
	if field.Optional || field.DefaultValue != nil || field.DefaultFunc != nil {
		return true
	}
	// update makes write-only fields optional so they can be left alone
	return sqlQuery == "update" && !field.CanApiRead()
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
	return snakeToCamelCase(name)
}

func toUnexportedName(name string) string {
	if name == "" {
		return name
	}
	return strings.ToLower(name[:1]) + name[1:]
}

func qualifyType(expr ast.Expr, pkg string) string {
	switch t := expr.(type) {
	case *ast.Ident:
		// Queries is re-declared locally as the wrapped type, so keep it local.
		if t.Name == "Queries" {
			return "Queries"
		}
		if ast.IsExported(t.Name) {
			return pkg + "." + t.Name
		}
		return t.Name
	case *ast.StarExpr:
		return "*" + qualifyType(t.X, pkg)
	case *ast.ArrayType:
		return "[]" + qualifyType(t.Elt, pkg)
	case *ast.Ellipsis:
		return "..." + qualifyType(t.Elt, pkg)
	case *ast.MapType:
		return "map[" + qualifyType(t.Key, pkg) + "]" + qualifyType(t.Value, pkg)
	case *ast.SelectorExpr:
		return formatType(t)
	default:
		return formatType(expr)
	}
}

// usesPackage reports whether body references the package selector name (as
// "name."), matching only at an identifier boundary so "time" does not match
// inside "runtime.".
func usesPackage(body, name string) bool {
	sel := name + "."
	from := 0
	for {
		i := strings.Index(body[from:], sel)
		if i < 0 {
			return false
		}
		pos := from + i
		if pos == 0 || !isIdentByte(body[pos-1]) {
			return true
		}
		from = pos + len(sel)
	}
}

func isIdentByte(c byte) bool {
	return c == '_' ||
		(c >= 'a' && c <= 'z') ||
		(c >= 'A' && c <= 'Z') ||
		(c >= '0' && c <= '9')
}

func sqlToGo(field schema.Field, pbFieldRef string, sqlDialect schema.SQLDialect) string {
	if sqlDialect == schema.SQLite {
		if field.Type == schema.FieldTypeBool {
			if field.Optional {
				return fmt.Sprintf("SQLiteBoolPtrToInt64Ptr(%s)", pbFieldRef)
			}
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

	if sqlDialect == schema.MySQL && field.Optional {
		if field.Type == schema.FieldTypeByte {
			// Special case: some generated refs may be dereferenced (e.g. *arg.ApiKey),
			// but PtrBytesToNullString expects a pointer, so strip a leading '*'.
			normalizedRef := strings.TrimPrefix(pbFieldRef, "*")
			return fmt.Sprintf("PtrBytesToNullString(%s)", normalizedRef)
		}
	}

	// SQLite/Postgres store bytes as []byte, which is already nilable, but the
	// wrapper keeps optional bytes as *[]byte for cross-dialect consistency.
	if field.Optional && field.Type == schema.FieldTypeByte &&
		(sqlDialect == schema.SQLite || sqlDialect == schema.PostgreSQL) {
		return fmt.Sprintf("PtrToNullBytes(%s)", pbFieldRef)
	}

	if field.Optional && (sqlDialect == schema.PostgreSQL || sqlDialect == schema.MySQL) {
		switch field.Type {
		case schema.FieldTypeString, schema.FieldTypeJSON:
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
func goFromSQL(field schema.Field, dbFieldRef string, sqlDialect schema.SQLDialect) string {
	if sqlDialect == schema.SQLite {
		if field.Type == schema.FieldTypeBool {
			if field.Optional {
				return fmt.Sprintf("SQLiteInt64PtrToBoolPtr(%s)", dbFieldRef)
			}
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

	// SQLite/Postgres return bytes as []byte; convert back to the wrapper's *[]byte.
	if field.Optional && field.Type == schema.FieldTypeByte &&
		(sqlDialect == schema.SQLite || sqlDialect == schema.PostgreSQL) {
		return fmt.Sprintf("NullBytesToPtr(%s)", dbFieldRef)
	}

	if field.Optional && (sqlDialect == schema.PostgreSQL || sqlDialect == schema.MySQL) {
		switch field.Type {
		case schema.FieldTypeString, schema.FieldTypeJSON:
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
