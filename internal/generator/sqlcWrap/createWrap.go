package sqlcwrap

import (
	"fmt"
	"go/ast"
	"strings"

	"github.com/guntisdev/entlite/internal/generator/sqlc"
	"github.com/guntisdev/entlite/internal/schema"
	"github.com/guntisdev/entlite/pkg/entlite/permissions"
)

func generateCreateStruct(structName string, structType *ast.StructType, entity schema.Entity) string {
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
			if field.DefaultFunc != nil {
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

func generateCreateMethod(funcDecl *ast.FuncDecl, entity schema.Entity, inputPkg string, sqlDialect sqlc.SQLDialect) string {
	var sb strings.Builder

	receiverType := formatType(funcDecl.Recv.List[0].Type)
	sb.WriteString(fmt.Sprintf("func (q %s) %s(ctx context.Context, arg %sParams) ", receiverType, funcDecl.Name.Name, funcDecl.Name.Name))

	var firstReturnType string
	idFieldPtr := entity.GetIdField()
	if idFieldPtr != nil {
		field := *idFieldPtr
		firstReturnType = string(field.Type)
	}

	// sqlc always generates (result, error)
	if funcDecl.Type.Results != nil && len(funcDecl.Type.Results.List) == 2 {
		if firstReturnType == "" {
			firstReturnType = formatType(funcDecl.Type.Results.List[0].Type)
		}
		secondReturnType := formatType(funcDecl.Type.Results.List[1].Type)
		sb.WriteString(fmt.Sprintf("(%s, %s)", firstReturnType, secondReturnType))
	}

	sb.WriteString(" {\n")
	sb.WriteString(addValidationChecks(entity, "create", firstReturnType))
	sb.WriteString(fmt.Sprintf("\tinternalArg := %s.%sParams{\n", inputPkg, funcDecl.Name.Name))

	defaultFuncFields := make(map[string]schema.Field)
	for _, field := range entity.Fields {
		if field.DefaultFunc != nil {
			defaultFuncFields[toExportedName(field.Name)] = field
		}
	}

	for _, field := range entity.Fields {
		exportedName := toExportedName(field.Name)
		if field.IsID() && field.DefaultFunc == nil {
			continue
		}
		if _, hasDefaultFunc := defaultFuncFields[exportedName]; hasDefaultFunc {
			funcName := field.DefaultFunc().(string)
			canApiWrite := (field.Permissions & permissions.ApiWrite) != 0
			if canApiWrite {
				convertField := sqlToGo(field, fmt.Sprintf("arg.%s", exportedName), sqlDialect)
				sb.WriteString(fmt.Sprintf("\t\t%s: OptionalWithFallback(%s, %s()),\n", exportedName, convertField, funcName))
			} else {
				sb.WriteString(fmt.Sprintf("\t\t%s: %s(),\n", exportedName, funcName))
			}
		} else {
			convertField := sqlToGo(field, fmt.Sprintf("arg.%s", exportedName), sqlDialect)
			sb.WriteString(fmt.Sprintf("\t\t%s: %s,\n", exportedName, convertField))
		}
	}

	sb.WriteString("\t}\n")

	// Handle return value conversion for SQLite ID (int64 -> int32)
	idField := entity.GetIdField()
	if (sqlDialect == sqlc.SQLite || sqlDialect == sqlc.MySQL) && idField != nil && idField.Type == schema.FieldTypeInt {
		sb.WriteString(fmt.Sprintf("\tid, err := (*%s.Queries)(q).%s(ctx, internalArg)\n", inputPkg, funcDecl.Name.Name))
		sb.WriteString("\treturn IntConvert[int64, int32](id), err\n")
	} else {
		sb.WriteString(fmt.Sprintf("\treturn (*%s.Queries)(q).%s(ctx, internalArg)\n", inputPkg, funcDecl.Name.Name))
	}

	sb.WriteString("}\n\n")

	return sb.String()
}
