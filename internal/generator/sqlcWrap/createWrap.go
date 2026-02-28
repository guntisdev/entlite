package sqlcwrap

import (
	"fmt"
	"go/ast"
	"strings"

	"github.com/guntisdev/entlite/internal/schema"
)

func generateCreateStruct(structName string, structType *ast.StructType, entity schema.Entity) string {
	var sb strings.Builder

	defaultFuncFields := make(map[string]bool)
	for _, field := range entity.Fields {
		if field.DefaultFunc != nil {
			defaultFuncFields[toExportedName(field.Name)] = true
		}
	}

	sb.WriteString(fmt.Sprintf("type %s struct {\n", structName))

	for _, field := range structType.Fields.List {
		if len(field.Names) > 0 {
			fieldName := field.Names[0].Name
			// Skip fields that have DefaultFunc
			if !defaultFuncFields[fieldName] {
				sb.WriteString(fmt.Sprintf("\t%s %s", fieldName, formatType(field.Type)))
				if field.Tag != nil {
					sb.WriteString(fmt.Sprintf(" %s", field.Tag.Value))
				}
				sb.WriteString("\n")
			}
		}
	}

	sb.WriteString("}\n\n")
	return sb.String()
}

func generateCreateMethod(funcDecl *ast.FuncDecl, entity schema.Entity, inputPkg string) string {
	var sb strings.Builder

	receiverType := formatType(funcDecl.Recv.List[0].Type)
	sb.WriteString(fmt.Sprintf("func (q %s) %s(ctx context.Context, arg %sParams) ", receiverType, funcDecl.Name.Name, funcDecl.Name.Name))

	if funcDecl.Type.Results != nil && len(funcDecl.Type.Results.List) > 0 {
		sb.WriteString("(")
		for i, result := range funcDecl.Type.Results.List {
			if i > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(formatType(result.Type))
		}
		sb.WriteString(")")
	}

	sb.WriteString(" {\n")

	idDefaultVal := "0"
	for _, field := range entity.Fields {
		if field.IsID() && field.Type == schema.FieldTypeString {
			idDefaultVal = "\"\""
		}
	}

	for _, field := range entity.Fields {
		if field.Validate == nil {
			continue
		}

		validateName := field.Validate().(string)
		fieldName := toDBFieldName(field)
		sb.WriteString(fmt.Sprintf("\tif !%s(arg.%s) {\n", validateName, fieldName))
		sb.WriteString(fmt.Sprintf("\t\treturn %s, fmt.Errorf(\"Failed create: incorrect value for %s in field %s, validated by %s\")\n", idDefaultVal, entity.Name, field.Name, validateName))
		sb.WriteString("\t}\n")
	}

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
			sb.WriteString(fmt.Sprintf("\t\t%s: %s(),\n", exportedName, funcName))
		} else {
			sb.WriteString(fmt.Sprintf("\t\t%s: arg.%s,\n", exportedName, exportedName))
		}
	}

	sb.WriteString("\t}\n")

	sb.WriteString(fmt.Sprintf("\treturn (*%s.Queries)(q).%s(ctx, internalArg)\n", inputPkg, funcDecl.Name.Name))
	sb.WriteString("}\n\n")

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
