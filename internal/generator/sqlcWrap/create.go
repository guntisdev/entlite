package sqlcwrap

import (
	"fmt"
	"go/ast"
	"strings"

	"github.com/guntisdev/entlite/internal/schema"
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
			canApiWrite := entity.CanFieldWrite(field)
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

	// without an id field the db generates nothing, so the insert only returns an error
	if !entity.HasIdField() {
		sb.WriteString("error {\n")
		sb.WriteString(addValidationChecks(entity, "create", errorOnlyReturn, "arg", "\t"))
		sb.WriteString(fmt.Sprintf("\tinternalArg := %s.%sParams{\n", inputPkg, funcDecl.Name.Name))
		writeCreateParamsFields(&sb, entity, "arg", "\t\t", sqlDialect)
		sb.WriteString("\t}\n")
		sb.WriteString(fmt.Sprintf("\treturn (*%s.Queries)(q).%s(ctx, internalArg)\n", inputPkg, funcDecl.Name.Name))
		sb.WriteString("}\n\n")

		return sb.String()
	}

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
		if field.IsVirtual() {
			continue
		}
		if field.IsID() && field.DefaultFunc == nil && field.DefaultValue == nil {
			continue
		}
		// a default, func or value, is a fallback: resolve the optional arg against it
		// first, then convert for the dialect
		if _, hasDefaultFunc := defaultFuncFields[exportedName]; hasDefaultFunc {
			funcName := field.DefaultFunc().(string)
			canApiWrite := entity.CanFieldWrite(field)
			if canApiWrite {
				fallbackRef := fmt.Sprintf("OptionalWithFallback(%s.%s, %s())", argVar, exportedName, funcName)
				sb.WriteString(fmt.Sprintf("%s%s: %s,\n", indent, exportedName, sqlToGo(field, fallbackRef, sqlDialect)))
			} else {
				sb.WriteString(fmt.Sprintf("%s%s: %s(),\n", indent, exportedName, funcName))
			}
		} else if defValField, hasDefaultVal := defaultValueFields[exportedName]; hasDefaultVal {
			valueLiteral := formatDefaultValue(defValField)
			canApiWrite := entity.CanFieldWrite(defValField)
			if canApiWrite {
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

// generateCreateBulkQuery wraps the sqlc single-row CreateBulk<Entity> insert.
func generateCreateBulkQuery(funcDecl *ast.FuncDecl, entity schema.Entity, inputPkg string, sqlDialect schema.SQLDialect) string {
	var sb strings.Builder

	receiverType := formatType(funcDecl.Recv.List[0].Type)
	queryName := funcDecl.Name.Name
	paramsType := queryName + "Params"
	internalParamsType := fmt.Sprintf("%s.%sParams", inputPkg, queryName)
	rowsFunc := toUnexportedName(queryName) + "Rows"

	// without an id field there is nothing to collect, every insert returns an error only
	hasID := entity.HasIdField()
	idType := ""
	resultsType := "error"
	nilResult := ""
	if hasID {
		idType = fieldToGoType(entity.GetIdField())
		resultsType = fmt.Sprintf("([]%s, error)", idType)
		nilResult = "nil, "
	}

	sb.WriteString(fmt.Sprintf("// %s inserts every row through q, which the caller binds to a transaction.\n", rowsFunc))
	sb.WriteString(fmt.Sprintf("func %s(ctx context.Context, q *%s.Queries, args []%s) %s {\n", rowsFunc, inputPkg, internalParamsType, resultsType))
	if hasID {
		// Handle return value conversion for SQLite/MySQL ID (int64 -> int32)
		idExpr := "id"
		if (sqlDialect == schema.SQLite || sqlDialect == schema.MySQL) && entity.GetIdField().Type == schema.FieldTypeInt {
			idExpr = "IntConvert[int64, int32](id)"
		}
		sb.WriteString(fmt.Sprintf("\tresults := make([]%s, 0, len(args))\n", idType))
		sb.WriteString("\tfor _, internalArg := range args {\n")
		sb.WriteString(fmt.Sprintf("\t\tid, err := q.%s(ctx, internalArg)\n", queryName))
		sb.WriteString("\t\tif err != nil {\n")
		sb.WriteString("\t\t\treturn nil, err\n")
		sb.WriteString("\t\t}\n")
		sb.WriteString(fmt.Sprintf("\t\tresults = append(results, %s)\n", idExpr))
		sb.WriteString("\t}\n")
		sb.WriteString("\treturn results, nil\n")
	} else {
		sb.WriteString("\tfor _, internalArg := range args {\n")
		sb.WriteString(fmt.Sprintf("\t\tif err := q.%s(ctx, internalArg); err != nil {\n", queryName))
		sb.WriteString("\t\t\treturn err\n")
		sb.WriteString("\t\t}\n")
		sb.WriteString("\t}\n")
		sb.WriteString("\treturn nil\n")
	}
	sb.WriteString("}\n\n")

	sb.WriteString(fmt.Sprintf("func (q %s) %s(ctx context.Context, args []%s) %s {\n", receiverType, queryName, paramsType, resultsType))
	sb.WriteString("\tif len(args) == 0 {\n")
	if hasID {
		sb.WriteString(fmt.Sprintf("\t\treturn []%s{}, nil\n", idType))
	} else {
		sb.WriteString("\t\treturn nil\n")
	}
	sb.WriteString("\t}\n\n")

	indexVar := "_"
	validationReturn := "nil"
	if !hasID {
		validationReturn = errorOnlyReturn
	}
	validation := addValidationChecksIndexed(entity, "create_bulk", validationReturn, "item", "\t\t", "i")
	if validation != "" {
		indexVar = "i"
	}
	sb.WriteString(fmt.Sprintf("\tinternalArgs := make([]%s, 0, len(args))\n", internalParamsType))
	sb.WriteString(fmt.Sprintf("\tfor %s, item := range args {\n", indexVar))
	sb.WriteString(validation)
	sb.WriteString(fmt.Sprintf("\t\tinternalArgs = append(internalArgs, %s{\n", internalParamsType))
	writeCreateParamsFields(&sb, entity, "item", "\t\t\t", sqlDialect)
	sb.WriteString("\t\t})\n")
	sb.WriteString("\t}\n\n")

	sb.WriteString(fmt.Sprintf("\tinternalQueries := (*%s.Queries)(q)\n", inputPkg))
	sb.WriteString("\tbeginner, ok := internalQueries.DB().(txBeginner)\n")
	sb.WriteString("\tif !ok {\n")
	sb.WriteString("\t\t// Already inside a transaction; the caller owns atomicity.\n")
	sb.WriteString(fmt.Sprintf("\t\treturn %s(ctx, internalQueries, internalArgs)\n", rowsFunc))
	sb.WriteString("\t}\n\n")

	sb.WriteString("\ttx, err := beginner.BeginTx(ctx, nil)\n")
	sb.WriteString("\tif err != nil {\n")
	sb.WriteString(fmt.Sprintf("\t\treturn %serr\n", nilResult))
	sb.WriteString("\t}\n")
	sb.WriteString("\tdefer func() { _ = tx.Rollback() }()\n\n")
	if hasID {
		sb.WriteString(fmt.Sprintf("\tresults, err := %s(ctx, internalQueries.WithTx(tx), internalArgs)\n", rowsFunc))
		sb.WriteString("\tif err != nil {\n")
		sb.WriteString("\t\treturn nil, err\n")
		sb.WriteString("\t}\n")
		sb.WriteString("\tif err := tx.Commit(); err != nil {\n")
		sb.WriteString("\t\treturn nil, err\n")
		sb.WriteString("\t}\n")
		sb.WriteString("\treturn results, nil\n")
	} else {
		sb.WriteString(fmt.Sprintf("\tif err := %s(ctx, internalQueries.WithTx(tx), internalArgs); err != nil {\n", rowsFunc))
		sb.WriteString("\t\treturn err\n")
		sb.WriteString("\t}\n")
		sb.WriteString("\treturn tx.Commit()\n")
	}
	sb.WriteString("}\n\n")

	return sb.String()
}
