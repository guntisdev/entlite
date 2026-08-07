package sqlcwrap

import (
	"fmt"
	"go/ast"
	"strings"

	"github.com/guntisdev/entlite/internal/schema"
)

func (ctx *generationContext) generateListQuery(funcDecl *ast.FuncDecl, entity schema.Entity) string {
	var sb strings.Builder
	inputPkg := ctx.inputPackageName

	params, args, prelude := ctx.wrapFilterParams(funcDecl, entity)

	receiverType := formatType(funcDecl.Recv.List[0].Type)
	sb.WriteString(fmt.Sprintf("func (q %s) %s(ctx context.Context%s) ", receiverType, funcDecl.Name.Name, params))

	if funcDecl.Type.Results != nil && len(funcDecl.Type.Results.List) == 2 {
		sb.WriteString(fmt.Sprintf("([]*%s, error)", entity.Name))
	}

	sb.WriteString(" {\n")
	sb.WriteString(prelude)

	sb.WriteString(fmt.Sprintf("\tdbResults, err := (*%s.Queries)(q).%s(ctx%s)\n", inputPkg, funcDecl.Name.Name, args))
	sb.WriteString("\tif err != nil {\n")
	sb.WriteString("\t\treturn nil, err\n")
	sb.WriteString("\t}\n")

	sb.WriteString(fmt.Sprintf("\tresult := make([]*%s, len(dbResults))\n", entity.Name))
	sb.WriteString("\tfor i := range dbResults {\n")
	sb.WriteString(fmt.Sprintf("\t\tresult[i] = %sFromSQL(&dbResults[i])\n", entity.Name))
	sb.WriteString("\t}\n")
	sb.WriteString("\treturn result, nil\n")
	sb.WriteString("}\n\n")

	return sb.String()
}
