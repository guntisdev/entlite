package parser

import (
	"fmt"
	"go/ast"
	"go/token"
	"strings"

	"github.com/guntisdev/entlite/internal/schema"
)

func parseQueriesMethod(funcDecl *ast.FuncDecl) ([]schema.Query, error) {
	var queries []schema.Query

	if funcDecl.Body == nil {
		return queries, nil
	}

	for _, stmt := range funcDecl.Body.List {
		retStmt, ok := stmt.(*ast.ReturnStmt)
		if !ok {
			continue
		}

		for _, result := range retStmt.Results {
			if compLit, ok := result.(*ast.CompositeLit); ok {
				for _, elt := range compLit.Elts {
					parsedQueries, err := parseQueryExpression(elt)
					if err != nil {
						return nil, err
					}
					queries = append(queries, parsedQueries...)
				}

			}
		}
	}

	return queries, nil
}

func parseQueryExpression(expr ast.Expr) ([]schema.Query, error) {
	callExpr, ok := expr.(*ast.CallExpr)
	if !ok {
		return nil, nil
	}

	queries, handled, err := parseQueryCall(callExpr)
	if err != nil {
		return nil, err
	}
	if !handled {
		return nil, nil
	}

	return queries, nil
}

func parseQueryCall(callExpr *ast.CallExpr) ([]schema.Query, bool, error) {
	selExpr, ok := callExpr.Fun.(*ast.SelectorExpr)
	if !ok {
		return nil, false, nil
	}

	if ident, ok := selExpr.X.(*ast.Ident); ok && ident.Name == "query" {
		switch selExpr.Sel.Name {
		case "DefaultCRUD":
			return []schema.Query{
				{Type: schema.QueryCreate, Fields: []string{"ID"}},
				{Type: schema.QueryGetBy, Fields: []string{"ID"}},
				{Type: schema.QueryUpdate, Fields: []string{"ID"}},
				{Type: schema.QueryDelete, Fields: []string{"ID"}},
				{Type: schema.QueryListBy, Fields: []string{"ID"}},
			}, true, nil
		case "Create":
			return []schema.Query{{Type: schema.QueryCreate, Fields: []string{"ID"}}}, true, nil
		case "Get":
			return []schema.Query{{Type: schema.QueryGetBy, Fields: []string{"ID"}}}, true, nil
		case "Update":
			return []schema.Query{{Type: schema.QueryUpdate, Fields: []string{"ID"}}}, true, nil
		case "Delete":
			return []schema.Query{{Type: schema.QueryDelete, Fields: []string{"ID"}}}, true, nil
		case "List":
			return []schema.Query{{Type: schema.QueryListBy, Fields: []string{"ID"}}}, true, nil
		case "GetBy":
			fields, err := parseStringArgs(callExpr.Args)
			if err != nil {
				return nil, true, fmt.Errorf("GetBy expects string field args: %w", err)
			}
			return []schema.Query{{Type: schema.QueryGetBy, Fields: fields}}, true, nil
		case "ListBy":
			fields, filters, err := parseListByArgs(callExpr.Args)
			if err != nil {
				return nil, true, err
			}
			return []schema.Query{{Type: schema.QueryListBy, Fields: fields, Filters: filters}}, true, nil
		default:
			return nil, false, nil
		}
	}

	innerCall, ok := selExpr.X.(*ast.CallExpr)
	if !ok {
		return nil, false, nil
	}

	queries, handled, err := parseQueryCall(innerCall)
	if err != nil || !handled {
		return queries, handled, err
	}
	if len(queries) != 1 {
		return nil, true, fmt.Errorf("chained query operation %q only supported for a single query", selExpr.Sel.Name)
	}

	query := queries[0]
	if query.Type != schema.QueryListBy {
		return nil, true, fmt.Errorf("%s is only supported for ListBy queries", selExpr.Sel.Name)
	}

	switch selExpr.Sel.Name {
	case "Count":
		if len(callExpr.Args) != 0 {
			return nil, true, fmt.Errorf("Count does not accept arguments")
		}
		query.Count = true
	case "OrderBy":
		if len(callExpr.Args) != 1 {
			return nil, true, fmt.Errorf("OrderBy expects exactly one string field")
		}
		orderField, err := parseSingleStringArg(callExpr.Args[0])
		if err != nil {
			return nil, true, fmt.Errorf("OrderBy expects exactly one string field: %w", err)
		}
		query.OrderBy = orderField
	default:
		return nil, false, nil
	}

	return []schema.Query{query}, true, nil
}

func parseStringArgs(args []ast.Expr) ([]string, error) {
	fields := make([]string, 0, len(args))
	for _, arg := range args {
		field, err := parseSingleStringArg(arg)
		if err != nil {
			return nil, err
		}
		fields = append(fields, field)
	}

	return fields, nil
}

func parseSingleStringArg(arg ast.Expr) (string, error) {
	lit, ok := arg.(*ast.BasicLit)
	if !ok || lit.Kind != token.STRING {
		return "", fmt.Errorf("expected string literal")
	}

	return strings.Trim(lit.Value, "\""), nil
}

func parseListByArgs(args []ast.Expr) ([]string, []schema.QueryFilter, error) {
	fields := []string{}
	filters := []schema.QueryFilter{}

	for _, arg := range args {
		if field, err := parseSingleStringArg(arg); err == nil {
			fields = append(fields, field)
			continue
		}

		parsedFilter, ok, err := parseFilterExpression(arg)
		if err != nil {
			return nil, nil, err
		}
		if ok {
			filters = append(filters, parsedFilter)
			continue
		}

		return nil, nil, fmt.Errorf("ListBy argument must be either a string field or filter.* call")
	}

	if len(fields) > 0 && len(filters) > 0 {
		return nil, nil, fmt.Errorf("ListBy accepts either string fields or filters, but not both")
	}

	return fields, filters, nil
}

func parseFilterExpression(expr ast.Expr) (schema.QueryFilter, bool, error) {
	callExpr, ok := expr.(*ast.CallExpr)
	if !ok {
		return schema.QueryFilter{}, false, nil
	}

	selExpr, ok := callExpr.Fun.(*ast.SelectorExpr)
	if !ok {
		return schema.QueryFilter{}, false, nil
	}

	if selExpr.Sel.Name == "Optional" {
		innerCall, ok := selExpr.X.(*ast.CallExpr)
		if !ok {
			return schema.QueryFilter{}, true, fmt.Errorf("Optional must be chained from a filter call")
		}
		parsedFilter, handled, err := parseFilterExpression(innerCall)
		if err != nil {
			return schema.QueryFilter{}, true, err
		}
		if !handled {
			return schema.QueryFilter{}, true, fmt.Errorf("Optional must be chained from filter.Range/filter.Search/filter.Eq")
		}
		if len(callExpr.Args) != 0 {
			return schema.QueryFilter{}, true, fmt.Errorf("Optional does not accept arguments")
		}

		parsedFilter.Optional = true
		return parsedFilter, true, nil
	}

	ident, ok := selExpr.X.(*ast.Ident)
	if !ok || ident.Name != "filter" {
		return schema.QueryFilter{}, false, nil
	}

	if len(callExpr.Args) != 1 {
		return schema.QueryFilter{}, true, fmt.Errorf("filter.%s expects exactly one string field", selExpr.Sel.Name)
	}

	field, err := parseSingleStringArg(callExpr.Args[0])
	if err != nil {
		return schema.QueryFilter{}, true, fmt.Errorf("filter.%s expects exactly one string field", selExpr.Sel.Name)
	}

	parsedFilter := schema.QueryFilter{Field: field}
	switch selExpr.Sel.Name {
	case "Range":
		parsedFilter.Type = schema.QueryFilterRange
	case "Search":
		parsedFilter.Type = schema.QueryFilterSearch
	case "Eq":
		parsedFilter.Type = schema.QueryFilterEq
	default:
		return schema.QueryFilter{}, true, fmt.Errorf("unsupported filter function filter.%s", selExpr.Sel.Name)
	}

	return parsedFilter, true, nil
}

func validateQueryFields(entity schema.Entity) error {
	if len(entity.Queries) == 0 {
		return nil
	}

	for _, query := range entity.Queries {
		switch query.Type {
		case schema.QueryGetBy:
			if len(query.Fields) == 0 {
				return fmt.Errorf("entity %q has query %q with empty fields", entity.Name, query.Type)
			}

			for _, fieldName := range query.Fields {
				if !entityHasField(entity, fieldName) {
					return fmt.Errorf("entity %q query %q references nonexisting field %q", entity.Name, query.Type, fieldName)
				}
			}
		case schema.QueryListBy:
			if len(query.Fields) > 0 && len(query.Filters) > 0 {
				return fmt.Errorf("entity %q query %q mixes fields and filters; choose one", entity.Name, query.Type)
			}

			if len(query.Fields) == 0 && len(query.Filters) == 0 {
				return fmt.Errorf("entity %q has query %q with empty fields/filters", entity.Name, query.Type)
			}

			for _, fieldName := range query.Fields {
				if !entityHasField(entity, fieldName) {
					return fmt.Errorf("entity %q query %q references nonexisting field %q", entity.Name, query.Type, fieldName)
				}
			}

			for _, queryFilter := range query.Filters {
				if !entityHasField(entity, queryFilter.Field) {
					return fmt.Errorf("entity %q query %q filter references nonexisting field %q", entity.Name, query.Type, queryFilter.Field)
				}
			}

			if query.OrderBy != "" && !entityHasField(entity, query.OrderBy) {
				return fmt.Errorf("entity %q query %q order_by references nonexisting field %q", entity.Name, query.Type, query.OrderBy)
			}
		}
	}

	return nil
}

func entityHasField(entity schema.Entity, fieldName string) bool {
	for _, field := range entity.Fields {
		if strings.EqualFold(field.Name, fieldName) {
			return true
		}
	}

	return false
}
