package sqlcwrap

import (
	"fmt"
	"go/ast"
	"strings"

	"github.com/guntisdev/entlite/internal/schema"
)

func (ctx *generationContext) generateDeleteQuery(funcDecl *ast.FuncDecl, entity schema.Entity) string {
	var sb strings.Builder
	inputPkg := ctx.inputPackageName

	// The primary key params need no special case: they resolve to the entity fields
	params, args, prelude := ctx.wrapFilterParams(funcDecl, entity)

	receiverType := formatType(funcDecl.Recv.List[0].Type)
	sb.WriteString(fmt.Sprintf("func (q %s) %s(ctx context.Context%s) error {\n", receiverType, funcDecl.Name.Name, params))
	sb.WriteString(prelude)
	sb.WriteString(fmt.Sprintf("\treturn (*%s.Queries)(q).%s(ctx%s)\n", inputPkg, funcDecl.Name.Name, args))
	sb.WriteString("}\n\n")

	return sb.String()
}

func generateDeleteAllQuery(funcDecl *ast.FuncDecl, entity schema.Entity, inputPkg string, sqlDialect schema.SQLDialect) string {
	var sb strings.Builder

	receiverType := formatType(funcDecl.Recv.List[0].Type)
	sb.WriteString(fmt.Sprintf("func (q %s) %s(ctx context.Context) error {\n", receiverType, funcDecl.Name.Name))
	sb.WriteString(fmt.Sprintf("\treturn (*%s.Queries)(q).%s(ctx)\n", inputPkg, funcDecl.Name.Name))
	sb.WriteString("}\n\n")

	return sb.String()
}
