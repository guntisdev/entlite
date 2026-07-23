package parser

import (
	"fmt"
	"go/ast"

	"github.com/guntisdev/entlite/internal/schema"
)

func parseIndexesMethod(funcDecl *ast.FuncDecl) ([]schema.Index, error) {
	var indexes []schema.Index

	if funcDecl.Body == nil {
		return indexes, nil
	}

	for _, stmt := range funcDecl.Body.List {
		retStmt, ok := stmt.(*ast.ReturnStmt)
		if !ok {
			continue
		}

		for _, result := range retStmt.Results {
			if compLit, ok := result.(*ast.CompositeLit); ok {
				for _, elt := range compLit.Elts {
					index, handled, err := parseIndexExpression(elt)
					if err != nil {
						return nil, err
					}
					if handled {
						indexes = append(indexes, index)
					}
				}
			}
		}
	}

	return indexes, nil
}

func parseIndexExpression(expr ast.Expr) (schema.Index, bool, error) {
	callExpr, ok := expr.(*ast.CallExpr)
	if !ok {
		return schema.Index{}, false, nil
	}

	return parseIndexCall(callExpr)
}

func parseIndexCall(callExpr *ast.CallExpr) (schema.Index, bool, error) {
	selExpr, ok := callExpr.Fun.(*ast.SelectorExpr)
	if !ok {
		return schema.Index{}, false, nil
	}

	// Base constructors: index.Primary(...) / index.Fields(...)
	if ident, ok := selExpr.X.(*ast.Ident); ok && ident.Name == "index" {
		switch selExpr.Sel.Name {
		case "Primary":
			fields, err := parseStringArgs(callExpr.Args)
			if err != nil {
				return schema.Index{}, true, fmt.Errorf("index.Primary expects string field args: %w", err)
			}
			if len(fields) == 0 {
				return schema.Index{}, true, fmt.Errorf("index.Primary requires at least one field")
			}
			return schema.Index{Type: schema.IndexPrimary, Columns: columnsFromFields(fields)}, true, nil
		case "Fields":
			fields, err := parseStringArgs(callExpr.Args)
			if err != nil {
				return schema.Index{}, true, fmt.Errorf("index.Fields expects string field args: %w", err)
			}
			if len(fields) == 0 {
				return schema.Index{}, true, fmt.Errorf("index.Fields requires at least one field")
			}
			return schema.Index{Type: schema.IndexRegular, Columns: columnsFromFields(fields)}, true, nil
		default:
			return schema.Index{}, false, nil
		}
	}

	// Chained modifiers: <inner>.Unique() / .Name(..) / .Asc(..) / .Desc(..)
	innerCall, ok := selExpr.X.(*ast.CallExpr)
	if !ok {
		return schema.Index{}, false, nil
	}

	index, handled, err := parseIndexCall(innerCall)
	if err != nil || !handled {
		return index, handled, err
	}

	if index.Type == schema.IndexPrimary {
		return schema.Index{}, true, fmt.Errorf("%s cannot be chained on index.Primary", selExpr.Sel.Name)
	}

	switch selExpr.Sel.Name {
	case "Unique":
		if len(callExpr.Args) != 0 {
			return schema.Index{}, true, fmt.Errorf("Unique does not accept arguments")
		}
		index.Unique = true
	case "Name":
		if len(callExpr.Args) != 1 {
			return schema.Index{}, true, fmt.Errorf("Name expects exactly one string argument")
		}
		name, err := parseSingleStringArg(callExpr.Args[0])
		if err != nil {
			return schema.Index{}, true, fmt.Errorf("Name expects exactly one string argument: %w", err)
		}
		index.Name = name
	case "Asc":
		field, err := parseIndexColumnArg(callExpr.Args, "Asc")
		if err != nil {
			return schema.Index{}, true, err
		}
		index.Columns = append(index.Columns, schema.IndexColumn{Name: field, Desc: false})
	case "Desc":
		field, err := parseIndexColumnArg(callExpr.Args, "Desc")
		if err != nil {
			return schema.Index{}, true, err
		}
		index.Columns = append(index.Columns, schema.IndexColumn{Name: field, Desc: true})
	default:
		return schema.Index{}, true, fmt.Errorf("unsupported index operation %q", selExpr.Sel.Name)
	}

	return index, true, nil
}

func parseIndexColumnArg(args []ast.Expr, method string) (string, error) {
	if len(args) != 1 {
		return "", fmt.Errorf("%s expects exactly one string field", method)
	}
	field, err := parseSingleStringArg(args[0])
	if err != nil {
		return "", fmt.Errorf("%s expects exactly one string field: %w", method, err)
	}
	return field, nil
}

func columnsFromFields(fields []string) []schema.IndexColumn {
	cols := make([]schema.IndexColumn, len(fields))
	for i, f := range fields {
		cols[i] = schema.IndexColumn{Name: f}
	}
	return cols
}
