package sqlcwrap

import (
	"fmt"
	"go/ast"
	"strings"

	"github.com/guntisdev/entlite/internal/generator/sqlc"
	"github.com/guntisdev/entlite/internal/schema"
)

func generateListMethod(funcDecl *ast.FuncDecl, entity schema.Entity, inputPkg string, sqlDialect sqlc.SQLDialect) string {
	var sb strings.Builder

	receiverType := formatType(funcDecl.Recv.List[0].Type)
	sb.WriteString(fmt.Sprintf("func (q %s) %s(ctx context.Context", receiverType, funcDecl.Name.Name))

	if funcDecl.Type.Params != nil && len(funcDecl.Type.Params.List) > 1 {
		for i := 1; i < len(funcDecl.Type.Params.List); i++ {
			param := funcDecl.Type.Params.List[i]
			for _, name := range param.Names {
				sb.WriteString(fmt.Sprintf(", %s %s", name.Name, formatType(param.Type)))
			}
		}
	}

	sb.WriteString(") ")

	if funcDecl.Type.Results != nil && len(funcDecl.Type.Results.List) == 2 {
		sb.WriteString(fmt.Sprintf("([]*%s, error)", entity.Name))
	}

	sb.WriteString(" {\n")

	sb.WriteString(fmt.Sprintf("\tdbResults, err := (*%s.Queries)(q).%s(ctx", inputPkg, funcDecl.Name.Name))

	if funcDecl.Type.Params != nil && len(funcDecl.Type.Params.List) > 1 {
		for i := 1; i < len(funcDecl.Type.Params.List); i++ {
			param := funcDecl.Type.Params.List[i]
			for _, name := range param.Names {
				sb.WriteString(fmt.Sprintf(", %s", name.Name))
			}
		}
	}

	sb.WriteString(")\n")
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
