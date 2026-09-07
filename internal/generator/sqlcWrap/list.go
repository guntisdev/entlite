package sqlcwrap

import (
	"fmt"
	"go/ast"
	"strings"

	"github.com/guntisdev/entlite/internal/schema"
)

const totalSizeField = "TotalSize"

func (ctx *generationContext) generateListQuery(funcDecl *ast.FuncDecl, entity schema.Entity) string {
	var sb strings.Builder
	inputPkg := ctx.inputPackageName

	withCount := false
	if target, ok := ctx.dslQueries[funcDecl.Name.Name]; ok {
		withCount = target.query.Count
	}

	params, args, prelude := ctx.wrapFilterParams(funcDecl, entity)

	receiverType := formatType(funcDecl.Recv.List[0].Type)
	sb.WriteString(fmt.Sprintf("func (q %s) %s(ctx context.Context%s) ", receiverType, funcDecl.Name.Name, params))

	if funcDecl.Type.Results != nil && len(funcDecl.Type.Results.List) == 2 {
		if withCount {
			sb.WriteString(fmt.Sprintf("([]*%s, int64, error)", entity.Name))
		} else {
			sb.WriteString(fmt.Sprintf("([]*%s, error)", entity.Name))
		}
	}

	sb.WriteString(" {\n")
	sb.WriteString(prelude)

	errReturn := "\t\treturn nil, err\n"
	if withCount {
		errReturn = "\t\treturn nil, 0, err\n"
	}

	sb.WriteString(fmt.Sprintf("\tdbResults, err := (*%s.Queries)(q).%s(ctx%s)\n", inputPkg, funcDecl.Name.Name, args))
	sb.WriteString("\tif err != nil {\n")
	sb.WriteString(errReturn)
	sb.WriteString("\t}\n")

	sb.WriteString(fmt.Sprintf("\tresult := make([]*%s, len(dbResults))\n", entity.Name))
	sb.WriteString("\tfor i := range dbResults {\n")
	if withCount {
		sb.WriteString(ctx.rowToEntity(entity))
	} else {
		sb.WriteString(fmt.Sprintf("\t\tresult[i] = %sFromSQL(&dbResults[i])\n", entity.Name))
	}
	sb.WriteString("\t}\n")

	if !withCount {
		sb.WriteString("\treturn result, nil\n")
		sb.WriteString("}\n\n")
		return sb.String()
	}

	sb.WriteString("\tvar totalSize int64\n")
	sb.WriteString("\tif len(dbResults) > 0 {\n")
	sb.WriteString(fmt.Sprintf("\t\ttotalSize = dbResults[0].%s\n", totalSizeField))
	sb.WriteString("\t}\n")
	sb.WriteString("\treturn result, totalSize, nil\n")
	sb.WriteString("}\n\n")

	return sb.String()
}

func (ctx *generationContext) rowToEntity(entity schema.Entity) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("\t\tresult[i] = &%s{\n", entity.Name))
	for _, field := range entity.Fields {
		if field.IsVirtual() {
			continue
		}

		fieldName := toDBFieldName(field)
		convertedValue := goFromSQL(field, fmt.Sprintf("dbResults[i].%s", fieldName), ctx.sqlDialect)
		sb.WriteString(fmt.Sprintf("\t\t\t%s: %s,\n", fieldName, convertedValue))
	}
	sb.WriteString("\t\t}\n")

	return sb.String()
}
