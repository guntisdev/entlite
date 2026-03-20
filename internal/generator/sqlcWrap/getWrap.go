package sqlcwrap

import (
	"fmt"
	"go/ast"
	"strings"

	"github.com/guntisdev/entlite/internal/generator/sqlc"
	"github.com/guntisdev/entlite/internal/schema"
)

func generateGetMethod(funcDecl *ast.FuncDecl, entity schema.Entity, inputPkg string, sqlDialect sqlc.SQLDialect) string {
	var sb strings.Builder

	receiverType := formatType(funcDecl.Recv.List[0].Type)
	sb.WriteString(fmt.Sprintf("func (q %s) %s(ctx context.Context", receiverType, funcDecl.Name.Name))

	if funcDecl.Type.Params != nil && len(funcDecl.Type.Params.List) > 1 {
		for i := 1; i < len(funcDecl.Type.Params.List); i++ {
			param := funcDecl.Type.Params.List[i]
			for _, name := range param.Names {
				if strings.ToLower(name.Name) == "id" {
					idField := entity.GetIdField()
					if idField != nil {
						sb.WriteString(fmt.Sprintf(", %s %s", name.Name, fieldToGoType(*idField)))
					} else {
						sb.WriteString(fmt.Sprintf(", %s %s", name.Name, formatType(param.Type)))
					}
				} else {
					sb.WriteString(fmt.Sprintf(", %s %s", name.Name, formatType(param.Type)))
				}
			}
		}
	}

	sb.WriteString(") ")

	if funcDecl.Type.Results != nil && len(funcDecl.Type.Results.List) == 2 {
		sb.WriteString(fmt.Sprintf("(*pb.%s, error)", entity.Name))
	}

	sb.WriteString(" {\n")

	sb.WriteString(fmt.Sprintf("\tdbResult, err := (*%s.Queries)(q).%s(ctx", inputPkg, funcDecl.Name.Name))

	if funcDecl.Type.Params != nil && len(funcDecl.Type.Params.List) > 1 {
		for i := 1; i < len(funcDecl.Type.Params.List); i++ {
			param := funcDecl.Type.Params.List[i]
			for _, name := range param.Names {
				if strings.ToLower(name.Name) == "id" {
					idField := entity.GetIdField()
					if idField != nil && sqlDialect == sqlc.SQLite && idField.Type == schema.FieldTypeInt {
						// TODO use field converter
						sb.WriteString(", SQLiteInt32ToInt64(id)")
					} else {
						sb.WriteString(fmt.Sprintf(", %s", name.Name))
					}
				} else {
					sb.WriteString(fmt.Sprintf(", %s", name.Name))
				}
			}
		}
	}

	sb.WriteString(")\n")
	sb.WriteString("\tif err != nil {\n")
	sb.WriteString("\t\treturn nil, err\n")
	sb.WriteString("\t}\n")

	sb.WriteString(fmt.Sprintf("\treturn %sDBToProto(&dbResult), nil\n", entity.Name))
	sb.WriteString("}\n\n")

	return sb.String()
}

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
		sb.WriteString(fmt.Sprintf("([]*pb.%s, error)", entity.Name))
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

	sb.WriteString(fmt.Sprintf("\tresult := make([]*pb.%s, len(dbResults))\n", entity.Name))
	sb.WriteString("\tfor i := range dbResults {\n")
	sb.WriteString(fmt.Sprintf("\t\tresult[i] = %sDBToProto(&dbResults[i])\n", entity.Name))
	sb.WriteString("\t}\n")
	sb.WriteString("\treturn result, nil\n")
	sb.WriteString("}\n\n")

	return sb.String()
}

func generateDeleteMethod(funcDecl *ast.FuncDecl, entity schema.Entity, inputPkg string, sqlDialect sqlc.SQLDialect) string {
	var sb strings.Builder

	receiverType := formatType(funcDecl.Recv.List[0].Type)
	sb.WriteString(fmt.Sprintf("func (q %s) %s(ctx context.Context", receiverType, funcDecl.Name.Name))

	if funcDecl.Type.Params != nil && len(funcDecl.Type.Params.List) > 1 {
		for i := 1; i < len(funcDecl.Type.Params.List); i++ {
			param := funcDecl.Type.Params.List[i]
			for _, name := range param.Names {
				if strings.ToLower(name.Name) == "id" {
					idField := entity.GetIdField()
					if idField != nil {
						sb.WriteString(fmt.Sprintf(", %s %s", name.Name, fieldToGoType(*idField)))
					} else {
						sb.WriteString(fmt.Sprintf(", %s %s", name.Name, formatType(param.Type)))
					}
				} else {
					sb.WriteString(fmt.Sprintf(", %s %s", name.Name, formatType(param.Type)))
				}
			}
		}
	}

	sb.WriteString(") error {\n")

	sb.WriteString(fmt.Sprintf("\treturn (*%s.Queries)(q).%s(ctx", inputPkg, funcDecl.Name.Name))

	if funcDecl.Type.Params != nil && len(funcDecl.Type.Params.List) > 1 {
		for i := 1; i < len(funcDecl.Type.Params.List); i++ {
			param := funcDecl.Type.Params.List[i]
			for _, name := range param.Names {
				if strings.ToLower(name.Name) == "id" {
					idField := entity.GetIdField()
					if idField != nil && sqlDialect == sqlc.SQLite && idField.Type == schema.FieldTypeInt {
						// TODO use field converter
						sb.WriteString(", SQLiteInt32ToInt64(id)")
					} else {
						sb.WriteString(fmt.Sprintf(", %s", name.Name))
					}
				} else {
					sb.WriteString(fmt.Sprintf(", %s", name.Name))
				}
			}
		}
	}

	sb.WriteString(")\n")
	sb.WriteString("}\n\n")

	return sb.String()
}
