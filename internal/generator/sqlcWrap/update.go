package sqlcwrap

import (
	"fmt"
	"go/ast"
	"strings"

	"github.com/guntisdev/entlite/internal/schema"
	"github.com/guntisdev/entlite/internal/util"
)

func generateUpdateStruct(structName string, structType *ast.StructType, entity schema.Entity) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("type %s struct {\n", structName))

	for _, astField := range structType.Fields.List {
		if len(astField.Names) > 0 {
			fieldName := astField.Names[0].Name
			fieldPtr := getFieldByName(entity, fieldName)
			if fieldPtr == nil {
				continue
			}
			field := *fieldPtr

			canApiWrite := entity.CanFieldWrite(field)
			if !canApiWrite {
				continue
			}

			// write-only fields, e.g. a password, are optional in update
			canApiRead := entity.CanFieldRead(field)
			if field.DefaultFunc != nil || field.DefaultValue != nil || !canApiRead {
				field.Optional = true
			}

			sb.WriteString(fmt.Sprintf("\t%s %s", fieldName, fieldToGoType(field)))
			if astField.Tag != nil {
				sb.WriteString(fmt.Sprintf(" %s", astField.Tag.Value))
			}
			sb.WriteString("\n")
		}
	}

	sb.WriteString("}\n\n")
	return sb.String()
}

func generateUpdateQuery(funcDecl *ast.FuncDecl, entity schema.Entity, inputPkg string, sqlDialect schema.SQLDialect) string {
	var sb strings.Builder

	receiverType := formatType(funcDecl.Recv.List[0].Type)
	sb.WriteString(fmt.Sprintf("func (q %s) %s(ctx context.Context, arg %sParams) ", receiverType, funcDecl.Name.Name, funcDecl.Name.Name))

	sb.WriteString(fmt.Sprintf("(*%s, error)", entity.Name))

	sb.WriteString(" {\n")
	sb.WriteString(addValidationChecks(entity, "update", "nil", "arg", "\t"))
	sb.WriteString(fmt.Sprintf("\tinternalArg := %s.%sParams{\n", inputPkg, funcDecl.Name.Name))

	defaultFuncFields := make(map[string]schema.Field)
	defaultValueFields := make(map[string]schema.Field)
	for _, field := range entity.Fields {
		if field.DefaultFunc != nil {
			defaultFuncFields[toExportedName(field.Name)] = field
		}
		if field.DefaultValue != nil {
			defaultValueFields[toExportedName(field.Name)] = field
		}
	}

	for _, field := range entity.Fields {
		exportedName := toExportedName(field.Name)
		if field.IsVirtual() {
			continue
		}
		// Skip immutable fields, the primary key is needed for the WHERE clause
		if field.Immutable && !entity.IsPrimaryKeyField(field) {
			continue
		}

		canApiWrite := entity.CanFieldWrite(field)
		// write-only fields, e.g. a password, are optional in update
		canApiRead := entity.CanFieldRead(field)
		if !canApiRead {
			field.Optional = true
		}
		pointerStr := ""
		// only MySQL's PtrBytesToNullString takes a pointer ref, it strips the '*' back
		// off; SQLite and Postgres take the *[]byte as-is
		if field.Type == schema.FieldTypeByte && sqlDialect == schema.MySQL {
			pointerStr = "*"
		}
		if _, hasDefaultFunc := defaultFuncFields[exportedName]; hasDefaultFunc {
			funcName := field.DefaultFunc().(string)
			if canApiWrite {
				field.Optional = true
				convertField := sqlToGo(field, fmt.Sprintf("%sarg.%s", pointerStr, exportedName), sqlDialect)
				sb.WriteString(fmt.Sprintf("\t\t%s: %s,\n", exportedName, convertField))
			} else {
				sb.WriteString(fmt.Sprintf("\t\t%s: %s(),\n", exportedName, funcName))
			}
		} else if _, hasDefaultVal := defaultValueFields[exportedName]; hasDefaultVal {
			if canApiWrite {
				field.Optional = true
				convertField := sqlToGo(field, fmt.Sprintf("%sarg.%s", pointerStr, exportedName), sqlDialect)
				sb.WriteString(fmt.Sprintf("\t\t%s: %s,\n", exportedName, convertField))
			} else {
				continue
			}
		} else {
			convertField := sqlToGo(field, fmt.Sprintf("arg.%s", exportedName), sqlDialect)
			sb.WriteString(fmt.Sprintf("\t\t%s: %s,\n", exportedName, convertField))
		}
	}

	sb.WriteString("\t}\n\n")

	if sqlDialect == schema.MySQL {
		sb.WriteString(fmt.Sprintf("\terr := (*%s.Queries)(q).%s(ctx, internalArg)\n", inputPkg, funcDecl.Name.Name))
		sb.WriteString("\tif err != nil {\n")
		sb.WriteString("\t\treturn nil, err\n")
		sb.WriteString("\t}\n")
		getName := util.GenEntityGetByPrimaryKeyName(entity)
		sb.WriteString(fmt.Sprintf("\tdb%s, err := (*%s.Queries)(q).%s(ctx, %s)\n", entity.Name, inputPkg, getName, getByPrimaryKeyArg(entity, inputPkg, getName)))
		sb.WriteString("\tif err != nil {\n")
		sb.WriteString("\t\treturn nil, err\n")
		sb.WriteString("\t}\n")
	} else {
		sb.WriteString(fmt.Sprintf("\tdb%s, err := (*%s.Queries)(q).%s(ctx, internalArg)\n", entity.Name, inputPkg, funcDecl.Name.Name))
		sb.WriteString("\tif err != nil {\n")
		sb.WriteString("\t\treturn nil, err\n")
		sb.WriteString("\t}\n")
	}

	sb.WriteString(fmt.Sprintf("\treturn %sFromSQL(&db%s), nil\n", entity.Name, entity.Name))
	sb.WriteString("}\n\n")

	return sb.String()
}

func getByPrimaryKeyArg(entity schema.Entity, inputPkg, getName string) string {
	keyFields := entity.PrimaryKeyFields()

	if len(keyFields) == 1 {
		return fmt.Sprintf("arg.%s", toDBFieldName(keyFields[0]))
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%s.%sParams{", inputPkg, getName))
	for i, field := range keyFields {
		if i > 0 {
			sb.WriteString(", ")
		}
		name := toDBFieldName(field)
		sb.WriteString(fmt.Sprintf("%s: arg.%s", name, name))
	}
	sb.WriteString("}")

	return sb.String()
}
