package sqlcwrap

import (
	"fmt"
	"go/ast"
	"strings"

	"github.com/guntisdev/entlite/internal/generator/sqlc"
	"github.com/guntisdev/entlite/internal/schema"
	"github.com/guntisdev/entlite/pkg/entlite/permissions"
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

			canApiWrite := (field.Permissions & permissions.ApiWrite) != 0
			if !canApiWrite {
				continue
			}

			// special case for psw etc - if not readable then no obligatory to update
			canApiRead := (field.Permissions & permissions.ApiRead) != 0
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

func generateUpdateMethod(funcDecl *ast.FuncDecl, entity schema.Entity, inputPkg string, sqlDialect sqlc.SQLDialect) string {
	var sb strings.Builder

	receiverType := formatType(funcDecl.Recv.List[0].Type)
	sb.WriteString(fmt.Sprintf("func (q %s) %s(ctx context.Context, arg %sParams) ", receiverType, funcDecl.Name.Name, funcDecl.Name.Name))

	sb.WriteString(fmt.Sprintf("(*%s, error)", entity.Name))

	sb.WriteString(" {\n")
	sb.WriteString(addValidationChecks(entity, "update", "nil"))
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
		// Skip immutable fields (except ID which is needed for WHERE clause)
		if field.Immutable && !field.IsID() {
			continue
		}

		canApiWrite := (field.Permissions & permissions.ApiWrite) != 0
		// special case for psw etc - if not readable then no obligatory to update
		canApiRead := (field.Permissions & permissions.ApiRead) != 0
		if !canApiRead {
			field.Optional = true
		}
		pointerStr := ""
		if field.Type == schema.FieldTypeByte {
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
				fmt.Printf("%s %s > %s \n", field.Name, sqlDialect, convertField)
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

	if sqlDialect == sqlc.MySQL {
		sb.WriteString(fmt.Sprintf("\terr := (*%s.Queries)(q).%s(ctx, internalArg)\n", inputPkg, funcDecl.Name.Name))
		sb.WriteString("\tif err != nil {\n")
		sb.WriteString("\t\treturn nil, err\n")
		sb.WriteString("\t}\n")
		sb.WriteString(fmt.Sprintf("\tdb%s, err := (*%s.Queries)(q).Get%s(ctx, arg.ID)\n", entity.Name, inputPkg, entity.Name))
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
