package sqlcwrap

import (
	"fmt"
	"go/ast"
	"strconv"
	"strings"

	"github.com/guntisdev/entlite/internal/naming"
	"github.com/guntisdev/entlite/internal/schema"
)

func GoTypeFor(fieldType schema.FieldType) string {
	return fieldToGoType(schema.Field{Type: fieldType})
}

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
func filterParamField(query schema.Query, entity schema.Entity, paramName string) (schema.Field, bool) {
	lookup := func(name string) (schema.Field, bool) {
		for _, field := range entity.Fields {
			if strings.EqualFold(toDBFieldName(field), name) {
				return field, true
			}
		}
		return schema.Field{}, false
	}

	// an Optional() filter is skippable even when its own field is a required column,
	// so the param needs a nullable Go type independent of field.Optional
	withFilterOptional := func(field schema.Field) schema.Field {
		for _, f := range query.Filters {
			if f.Optional && strings.EqualFold(f.Field, field.Name) {
				field.Optional = true
				break
			}
		}
		return field
	}

	if field, ok := lookup(paramName); ok {
		return withFilterOptional(field), true
	}

	for _, prefix := range []string{"Min", "Max"} {
		if len(paramName) > len(prefix) && strings.EqualFold(paramName[:len(prefix)], prefix) {
			if field, ok := lookup(paramName[len(prefix):]); ok {
				return withFilterOptional(field), true
			}
		}
	}

	return schema.Field{}, false
}

// LIMIT/OFFSET are not entity fields; the wrapper keeps them int32, like the proto request
func isPaginationParam(fieldName string) bool {
	return fieldName == "Limit" || fieldName == "Offset"
}

// generateFilterParamsStruct restates sqlc's "<Query>Params" in the wrapper's types,
// keeping sqlc's field names and json tags.
func generateFilterParamsStruct(structName string, structType *ast.StructType, entity schema.Entity, query schema.Query) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("type %s struct {\n", structName))

	for _, astField := range structType.Fields.List {
		if len(astField.Names) == 0 {
			continue
		}
		fieldName := astField.Names[0].Name

		goType := formatType(astField.Type)
		if field, ok := filterParamField(query, entity, fieldName); ok {
			goType = fieldToGoType(field)
		} else if isPaginationParam(fieldName) {
			goType = "int32"
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

// generateFilterParamsArg builds the params literal for sqlc, converting each field
// back to its dialect type.
func generateFilterParamsArg(structName string, structType *ast.StructType, entity schema.Entity, inputPkg, argVar string, sqlDialect schema.SQLDialect, query schema.Query) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("\tinternalArg := %s.%s{\n", inputPkg, structName))

	for _, astField := range structType.Fields.List {
		if len(astField.Names) == 0 {
			continue
		}
		fieldName := astField.Names[0].Name

		valueRef := fmt.Sprintf("%s.%s", argVar, fieldName)
		if field, ok := filterParamField(query, entity, fieldName); ok {
			valueRef = sqlToGo(field, valueRef, sqlDialect)
		} else if isPaginationParam(fieldName) && formatType(astField.Type) == "int64" {
			// sqlite widens LIMIT/OFFSET to int64
			valueRef = fmt.Sprintf("IntConvert[int32, int64](%s)", valueRef)
		}

		sb.WriteString(fmt.Sprintf("\t\t%s: %s,\n", fieldName, valueRef))
	}

	sb.WriteString("\t}\n")
	return sb.String()
}

// wrapFilterParams renders a get/list wrapper's params, the forwarded args and any
// statements needed before the call.
func (ctx *generationContext) wrapFilterParams(funcDecl *ast.FuncDecl, entity schema.Entity) (params, args, prelude string) {
	if funcDecl.Type.Params == nil {
		return "", "", ""
	}

	var paramsSb, argsSb, preludeSb strings.Builder

	// The DSL query, if any, carries the Filters that decide optionality below.
	query := ctx.dslQueries[funcDecl.Name.Name].query

	// Index 0 is ctx, which callers emit themselves.
	for i := 1; i < len(funcDecl.Type.Params.List); i++ {
		param := funcDecl.Type.Params.List[i]
		typeName := formatType(param.Type)

		for _, name := range param.Names {
			// A params struct the wrapper restates: take ours, convert to sqlc's.
			if structType, ok := ctx.filterParamsStructs[typeName]; ok {
				paramsSb.WriteString(fmt.Sprintf(", %s %s", name.Name, typeName))
				preludeSb.WriteString(generateFilterParamsArg(typeName, structType, entity, ctx.inputPackageName, name.Name, ctx.sqlDialect, query))
				argsSb.WriteString(", internalArg")
				continue
			}

			// A lone filter arrives as a bare scalar rather than a struct.
			if field, ok := filterParamField(query, entity, name.Name); ok {
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

const errorOnlyReturn = "error"

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

	zeroPrefix := zeroValue + ", "
	if returnType == errorOnlyReturn {
		zeroPrefix = ""
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
		if !entity.CanFieldWrite(field) {
			continue
		}
		// update skips immutable fields, so they are not in the params struct
		if sqlQuery == "update" && field.Immutable {
			continue
		}

		ref := fmt.Sprintf("%s.%s", argVar, toDBFieldName(field))
		cond := fmt.Sprintf("!json.Valid([]byte(%s))", ref)
		if isPointerParam(entity, field, sqlQuery) {
			cond = fmt.Sprintf("%s != nil && !json.Valid([]byte(*%s))", ref, ref)
		}
		sb.WriteString(fmt.Sprintf("%sif %s {\n", indent, cond))
		sb.WriteString(fmt.Sprintf("%s\treturn %sfmt.Errorf(\"Failed %s: %sinvalid json for '%s' in field '%s'\"%s)\n", indent, zeroPrefix, sqlQuery, itemPrefix, entity.Name, field.Name, itemArgs))
		sb.WriteString(fmt.Sprintf("%s}\n", indent))
	}

	for _, field := range entity.Fields {
		if field.Validate == nil {
			continue
		}
		if field.IsVirtual() {
			continue
		}
		// update skips immutable fields, so they are not in the params struct
		if sqlQuery == "update" && field.Immutable {
			continue
		}

		validateName := field.Validate().(string)
		fieldName := toDBFieldName(field)
		ref := fmt.Sprintf("%s.%s", argVar, fieldName)
		cond := fmt.Sprintf("!%s(%s)", validateName, ref)
		if isPointerParam(entity, field, sqlQuery) {
			// an omitted optional field skips validation rather than dereferencing a nil pointer
			cond = fmt.Sprintf("%s != nil && !%s(*%s)", ref, validateName, ref)
		}
		sb.WriteString(fmt.Sprintf("%sif %s {\n", indent, cond))
		sb.WriteString(fmt.Sprintf("%s\treturn %sfmt.Errorf(\"Failed %s: %sincorrect value for '%s' in field '%s', validated by '%s'\"%s)\n", indent, zeroPrefix, sqlQuery, itemPrefix, entity.Name, field.Name, validateName, itemArgs))
		sb.WriteString(fmt.Sprintf("%s}\n", indent))
	}
	return sb.String()
}

// match sqlc conversion - ID and CamelCase names
func toDBFieldName(field schema.Field) string {
	return naming.SqlcGoName(field.Name)
}

// toProtoFieldName matches protoc-gen-go, which applies no Go initialisms:
// sensor_id becomes SensorId, not SensorID.
func toProtoFieldName(field schema.Field) string {
	return naming.ProtocGoName(field.Name)
}

// params are pointers when the field is optional or gets a default
func isPointerParam(entity schema.Entity, field schema.Field, sqlQuery string) bool {
	if field.Optional || field.DefaultValue != nil || field.DefaultFunc != nil {
		return true
	}
	// write-only fields, e.g. a password, are optional in update
	return sqlQuery == "update" && !entity.CanFieldRead(field)
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
	return naming.SqlcGoName(name)
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

// usesPackage reports whether body references "name.", matching only at an identifier
// boundary so "time" does not match inside "runtime.".
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
			// a generated ref may be dereferenced, e.g. *arg.ApiKey, but PtrBytesToNullString
			// wants a pointer, so strip a leading '*'
			normalizedRef := strings.TrimPrefix(pbFieldRef, "*")
			return fmt.Sprintf("PtrBytesToNullString(%s)", normalizedRef)
		}
	}

	// SQLite and Postgres store bytes as []byte, already nilable, but the wrapper keeps
	// optional bytes as *[]byte across dialects
	if field.Optional && field.Type == schema.FieldTypeByte &&
		(sqlDialect == schema.SQLite || sqlDialect == schema.PostgreSQL) {
		return fmt.Sprintf("PtrToNullBytes(%s)", pbFieldRef)
	}

	if field.Type == schema.FieldTypeJSON && sqlDialect != schema.SQLite {
		if !field.Optional {
			return fmt.Sprintf("StringToRawMessage(%s)", pbFieldRef)
		}
		if sqlDialect == schema.PostgreSQL {
			return fmt.Sprintf("PtrToNullRawMessage(%s)", pbFieldRef)
		}
		// mysql: real Optional() got TEXT (jsonCheck); Default() stays JSON
		if field.DefaultValue == nil {
			return fmt.Sprintf("PtrToNullString(%s)", pbFieldRef)
		}
		return fmt.Sprintf("PtrToRawMessage(%s)", pbFieldRef)
	}

	if field.Optional && (sqlDialect == schema.PostgreSQL || sqlDialect == schema.MySQL) {
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
		case schema.FieldTypeTime:
			return fmt.Sprintf("PtrToNullTime(%s)", pbFieldRef)
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

	// MySQL stores optional bytes as sql.NullString; convert back to *[]byte.
	if field.Optional && field.Type == schema.FieldTypeByte && sqlDialect == schema.MySQL {
		return fmt.Sprintf("NullStringToPtrBytes(%s)", dbFieldRef)
	}

	if field.Type == schema.FieldTypeJSON && sqlDialect != schema.SQLite {
		if !field.Optional {
			return fmt.Sprintf("RawMessageToString(%s)", dbFieldRef)
		}
		if sqlDialect == schema.PostgreSQL {
			return fmt.Sprintf("NullRawMessageToPtr(%s)", dbFieldRef)
		}
		if field.DefaultValue == nil {
			return fmt.Sprintf("NullStringToPtr(%s)", dbFieldRef)
		}
		return fmt.Sprintf("RawMessageToPtr(%s)", dbFieldRef)
	}

	if field.Optional && (sqlDialect == schema.PostgreSQL || sqlDialect == schema.MySQL) {
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
		case schema.FieldTypeTime:
			return fmt.Sprintf("NullTimeToPtr(%s)", dbFieldRef)
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
