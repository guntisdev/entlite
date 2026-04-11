package parser

import (
	"go/ast"

	"github.com/guntisdev/entlite/internal/schema"
)

func parseAnnotationsMethod(funcDecl *ast.FuncDecl) ([]schema.Annotation, error) {
	var annotations []schema.Annotation

	if funcDecl.Body == nil {
		return annotations, nil
	}

	for _, stmt := range funcDecl.Body.List {
		retStmt, ok := stmt.(*ast.ReturnStmt)
		if !ok {
			continue
		}

		for _, result := range retStmt.Results {
			if compLit, ok := result.(*ast.CompositeLit); ok {
				for _, elt := range compLit.Elts {
					if callExpr, ok := elt.(*ast.CallExpr); ok {
						annotation := parseAnnotationCall(callExpr)
						if annotation.Type != "" {
							annotations = append(annotations, annotation)
						}
					}
				}
			}
		}
	}

	return annotations, nil
}

func parseAnnotationCall(callExpr *ast.CallExpr) schema.Annotation {
	var annotation schema.Annotation

	if selExpr, ok := callExpr.Fun.(*ast.SelectorExpr); ok {
		if ident, ok := selExpr.X.(*ast.Ident); ok && ident.Name == "entlite" {
			switch selExpr.Sel.Name {
			case "Message":
				annotation.Type = schema.AnnotationMessage
			case "GRPC":
				annotation.Type = schema.AnnotationGRPC

				var methods []schema.Method
				if len(callExpr.Args) > 0 {
					// TODO deprecated, remove after Queries implemented
					methods = parseServiceArguments(callExpr.Args)
				} else {
					methods = []schema.Method{
						schema.MethodCreate,
						schema.MethodGet,
						schema.MethodUpdate,
						schema.MethodDelete,
						schema.MethodList,
					}
				}
				annotation.Methods = methods
			}
		}
	}

	return annotation
}
