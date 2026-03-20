package sqlcwrap

import (
	"fmt"
	"go/ast"
	"strings"

	"github.com/guntisdev/entlite/internal/schema"
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
			if field.DefaultFunc != nil {
				continue
			}
			fmt.Printf("Field: %s\n", field.Name)
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

func generateUpdateMethod(funcDecl *ast.FuncDecl, entity schema.Entity, inputPkg string) string {
	var sb strings.Builder

	receiverType := formatType(funcDecl.Recv.List[0].Type)
	sb.WriteString(fmt.Sprintf("func (q %s) %s(ctx context.Context, arg %sParams) ", receiverType, funcDecl.Name.Name, funcDecl.Name.Name))

	// sqlc always generates (result, error)
	var firstReturnType string
	if funcDecl.Type.Results != nil && len(funcDecl.Type.Results.List) == 2 {
		firstReturnType = formatType(funcDecl.Type.Results.List[0].Type)
		secondReturnType := formatType(funcDecl.Type.Results.List[1].Type)
		sb.WriteString(fmt.Sprintf("(%s, %s)", firstReturnType, secondReturnType))
	}

	sb.WriteString(" {\n")
	sb.WriteString(addValidationChecks(entity, "update", firstReturnType))
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
		if field.Immutable {
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
