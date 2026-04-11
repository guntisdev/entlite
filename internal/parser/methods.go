package parser

import (
	"go/ast"
	"go/token"

	"github.com/guntisdev/entlite/internal/schema"
)

func parseServiceArguments(args []ast.Expr) []schema.Method {
	var methods []schema.Method

	for _, arg := range args {
		if callExpr, ok := arg.(*ast.CallExpr); ok {
			if selExpr, ok := callExpr.Fun.(*ast.SelectorExpr); ok {
				if ident, ok := selExpr.X.(*ast.Ident); ok && ident.Name == "service" {
					if selExpr.Sel.Name == "Methods" {
						if len(callExpr.Args) > 0 {
							methodMethods := parseMethodsArguments(callExpr.Args[0])
							methods = append(methods, methodMethods...)
						}
					}
				}
			}
		}
	}

	if len(methods) == 0 {
		methods = []schema.Method{
			schema.MethodCreate,
			schema.MethodGet,
			schema.MethodUpdate,
			schema.MethodDelete,
			schema.MethodList,
		}
	}

	return methods
}

func parseMethodsArguments(expr ast.Expr) []schema.Method {
	var methods []schema.Method

	if binExpr, ok := expr.(*ast.BinaryExpr); ok && binExpr.Op == token.OR {
		leftMethods := parseMethodsArguments(binExpr.X)
		rightMethods := parseMethodsArguments(binExpr.Y)
		methods = append(methods, leftMethods...)
		methods = append(methods, rightMethods...)
	} else if selExpr, ok := expr.(*ast.SelectorExpr); ok {
		if ident, ok := selExpr.X.(*ast.Ident); ok && ident.Name == "service" {
			switch selExpr.Sel.Name {
			case "MethodCreate":
				methods = append(methods, schema.MethodCreate)
			case "MethodGet":
				methods = append(methods, schema.MethodGet)
			case "MethodUpdate":
				methods = append(methods, schema.MethodUpdate)
			case "MethodDelete":
				methods = append(methods, schema.MethodDelete)
			case "MethodList":
				methods = append(methods, schema.MethodList)
			}
		}
	}

	return methods
}
