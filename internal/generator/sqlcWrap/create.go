package sqlcwrap

import (
	"fmt"
	"go/ast"
	"strings"

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
			if field.DefaultFunc != nil || field.DefaultValue != nil {
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

func generateCreateQuery(funcDecl *ast.FuncDecl, entity schema.Entity, inputPkg string, sqlDialect schema.SQLDialect) string {
	var sb strings.Builder

	receiverType := formatType(funcDecl.Recv.List[0].Type)
	sb.WriteString(fmt.Sprintf("func (q %s) %s(ctx context.Context, arg %sParams) ", receiverType, funcDecl.Name.Name, funcDecl.Name.Name))

	var firstReturnType string
	idField := entity.GetIdField()
	firstReturnType = string(idField.Type)

	// sqlc always generates (result, error)
	if funcDecl.Type.Results != nil && len(funcDecl.Type.Results.List) == 2 {
		if firstReturnType == "" {
			firstReturnType = formatType(funcDecl.Type.Results.List[0].Type)
		}
		secondReturnType := formatType(funcDecl.Type.Results.List[1].Type)
		sb.WriteString(fmt.Sprintf("(%s, %s)", firstReturnType, secondReturnType))
	}

	sb.WriteString(" {\n")
	sb.WriteString(addValidationChecks(entity, "create", firstReturnType, "arg", "\t"))
	sb.WriteString(fmt.Sprintf("\tinternalArg := %s.%sParams{\n", inputPkg, funcDecl.Name.Name))
	writeCreateParamsFields(&sb, entity, "arg", "\t\t", sqlDialect)
	sb.WriteString("\t}\n")

	// Handle return value conversion for SQLite ID (int64 -> int32)
	if (sqlDialect == schema.SQLite || sqlDialect == schema.MySQL) && idField.Type == schema.FieldTypeInt {
		sb.WriteString(fmt.Sprintf("\tid, err := (*%s.Queries)(q).%s(ctx, internalArg)\n", inputPkg, funcDecl.Name.Name))
		sb.WriteString("\treturn IntConvert[int64, int32](id), err\n")
	} else {
		sb.WriteString(fmt.Sprintf("\treturn (*%s.Queries)(q).%s(ctx, internalArg)\n", inputPkg, funcDecl.Name.Name))
	}

	sb.WriteString("}\n\n")

	return sb.String()
}

func writeCreateParamsFields(sb *strings.Builder, entity schema.Entity, argVar, indent string, sqlDialect schema.SQLDialect) {
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
		if field.IsID() && field.DefaultFunc == nil && field.DefaultValue == nil {
			continue
		}
		if _, hasDefaultFunc := defaultFuncFields[exportedName]; hasDefaultFunc {
			funcName := field.DefaultFunc().(string)
			canApiWrite := (field.Permissions & permissions.ApiWrite) != 0
			if canApiWrite {
				// Resolve the optional arg against the fallback first, then apply
				// any dialect conversion around the resulting non-pointer value.
				fallbackRef := fmt.Sprintf("OptionalWithFallback(%s.%s, %s())", argVar, exportedName, funcName)
				sb.WriteString(fmt.Sprintf("%s%s: %s,\n", indent, exportedName, sqlToGo(field, fallbackRef, sqlDialect)))
			} else {
				sb.WriteString(fmt.Sprintf("%s%s: %s(),\n", indent, exportedName, funcName))
			}
		} else if defValField, hasDefaultVal := defaultValueFields[exportedName]; hasDefaultVal {
			valueLiteral := formatDefaultValue(defValField)
			canApiWrite := (defValField.Permissions & permissions.ApiWrite) != 0
			if canApiWrite {
				// Resolve the optional arg against the fallback first, then apply
				// any dialect conversion around the resulting non-pointer value.
				fallbackRef := fmt.Sprintf("OptionalWithFallback(%s.%s, %s)", argVar, exportedName, valueLiteral)
				sb.WriteString(fmt.Sprintf("%s%s: %s,\n", indent, exportedName, sqlToGo(defValField, fallbackRef, sqlDialect)))
			} else {
				sb.WriteString(fmt.Sprintf("%s%s: %s,\n", indent, exportedName, valueLiteral))
			}
		} else {
			convertField := sqlToGo(field, fmt.Sprintf("%s.%s", argVar, exportedName), sqlDialect)
			sb.WriteString(fmt.Sprintf("%s%s: %s,\n", indent, exportedName, convertField))
		}
	}
}

// generateCreateBulkQuery wraps the sqlc single-row CreateBulk<Entity> insert in a loop
func generateCreateBulkQuery(funcDecl *ast.FuncDecl, entity schema.Entity, inputPkg string, sqlDialect schema.SQLDialect) string {
	var sb strings.Builder

	receiverType := formatType(funcDecl.Recv.List[0].Type)
	paramsType := funcDecl.Name.Name + "Params"
	idField := entity.GetIdField()
	idType := fieldToGoType(idField)

	sb.WriteString(fmt.Sprintf("func (q %s) %s(ctx context.Context, args []%s) ([]%s, error) {\n", receiverType, funcDecl.Name.Name, paramsType, idType))
	sb.WriteString(fmt.Sprintf("\tresults := make([]%s, 0, len(args))\n", idType))
	sb.WriteString("\tfor _, item := range args {\n")
	sb.WriteString(addValidationChecks(entity, "create_bulk", "nil", "item", "\t\t"))
	sb.WriteString(fmt.Sprintf("\t\tinternalArg := %s.%sParams{\n", inputPkg, funcDecl.Name.Name))
	writeCreateParamsFields(&sb, entity, "item", "\t\t\t", sqlDialect)
	sb.WriteString("\t\t}\n")
	sb.WriteString(fmt.Sprintf("\t\tid, err := (*%s.Queries)(q).%s(ctx, internalArg)\n", inputPkg, funcDecl.Name.Name))
	sb.WriteString("\t\tif err != nil {\n")
	sb.WriteString("\t\t\treturn nil, err\n")
	sb.WriteString("\t\t}\n")
	// Handle return value conversion for SQLite/MySQL ID (int64 -> int32)
	if (sqlDialect == schema.SQLite || sqlDialect == schema.MySQL) && idField.Type == schema.FieldTypeInt {
		sb.WriteString("\t\tresults = append(results, IntConvert[int64, int32](id))\n")
	} else {
		sb.WriteString("\t\tresults = append(results, id)\n")
	}
	sb.WriteString("\t}\n")
	sb.WriteString("\treturn results, nil\n")
	sb.WriteString("}\n\n")

	return sb.String()
}
