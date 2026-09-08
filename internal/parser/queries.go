package parser

import (
	"fmt"
	"go/ast"
	"go/token"
	"strconv"
	"strings"

	"github.com/guntisdev/entlite/internal/schema"
	"github.com/guntisdev/entlite/internal/util"
)

func parseQueriesMethod(funcDecl *ast.FuncDecl, comments commentLookup) ([]schema.Query, error) {
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
				prevEnd := compLit.Lbrace
				for _, elt := range compLit.Elts {
					parsedQueries, err := parseQueryExpression(elt)
					if err != nil {
						return nil, err
					}

					// one expression can yield several queries, DefaultCRUD gives four
					comment := comments.docAbove(elt.Pos(), prevEnd)
					prevEnd = elt.End()
					for i := range parsedQueries {
						parsedQueries[i].Comment = comment
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
				{Type: schema.QueryCreate},
				{Type: schema.QueryGetBy, PrimaryKey: true},
				{Type: schema.QueryUpdate, PrimaryKey: true},
				{Type: schema.QueryDelete, PrimaryKey: true},
			}, true, nil
		case "Create":
			return []schema.Query{{Type: schema.QueryCreate}}, true, nil
		case "CreateBulk":
			if len(callExpr.Args) != 0 {
				return nil, true, fmt.Errorf("CreateBulk does not accept arguments")
			}
			return []schema.Query{{Type: schema.QueryCreateBulk}}, true, nil
		case "Get":
			return []schema.Query{{Type: schema.QueryGetBy, PrimaryKey: true}}, true, nil
		case "Update":
			return []schema.Query{{Type: schema.QueryUpdate, PrimaryKey: true}}, true, nil
		case "Delete":
			return []schema.Query{{Type: schema.QueryDelete, PrimaryKey: true}}, true, nil
		case "DeleteAll":
			if len(callExpr.Args) != 0 {
				return nil, true, fmt.Errorf("DeleteAll does not accept arguments")
			}
			return []schema.Query{{Type: schema.QueryDeleteAll}}, true, nil
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
		case "ListAll":
			if len(callExpr.Args) != 0 {
				return nil, true, fmt.Errorf("ListAll does not accept arguments")
			}
			return []schema.Query{{Type: schema.QueryListAll}}, true, nil
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
	switch selExpr.Sel.Name {
	case "Name", "Contracts":
	case "Count", "Distinct", "Limit", "Offset", "Asc", "Desc":
		if !query.IsList() {
			return nil, true, fmt.Errorf("%s is only supported for list queries", selExpr.Sel.Name)
		}
	case "Upsert", "Ignore":
		if !query.IsCreate() {
			return nil, true, fmt.Errorf("%s is only supported for create queries", selExpr.Sel.Name)
		}
	default:
		return nil, true, fmt.Errorf("unsupported query operation %q", selExpr.Sel.Name)
	}

	switch selExpr.Sel.Name {
	case "Count":
		if len(callExpr.Args) != 0 {
			return nil, true, fmt.Errorf("Count does not accept arguments")
		}
		query.Count = true
	case "Distinct":
		fields, err := parseStringArgs(callExpr.Args)
		if err != nil {
			return nil, true, fmt.Errorf("Distinct expects string field args: %w", err)
		}
		if len(fields) == 0 {
			return nil, true, fmt.Errorf("Distinct expects at least one field name")
		}
		query.Distinct = fields
	case "Asc", "Desc":
		field, err := parseColumnArg(callExpr.Args, selExpr.Sel.Name)
		if err != nil {
			return nil, true, err
		}
		if field == "" {
			return nil, true, fmt.Errorf("%s expects a field name", selExpr.Sel.Name)
		}
		query.OrderBy = append(query.OrderBy, schema.OrderColumn{Name: field, Desc: selExpr.Sel.Name == "Desc"})
	case "Limit":
		if len(callExpr.Args) > 1 {
			return nil, true, fmt.Errorf("Limit expects no arguments or a single row count")
		}
		query.HasLimit = true
		if len(callExpr.Args) == 1 {
			rows, err := parseSingleIntArg(callExpr.Args[0])
			if err != nil {
				return nil, true, fmt.Errorf("Limit expects no arguments or a single row count: %w", err)
			}
			if rows < 1 {
				return nil, true, fmt.Errorf("Limit %d must be at least 1", rows)
			}
			query.Limit = rows
		}
	case "Offset":
		if len(callExpr.Args) != 0 {
			return nil, true, fmt.Errorf("Offset does not accept arguments, the caller sends the value")
		}
		query.HasOffset = true
	case "Upsert":
		fields, err := parseStringArgs(callExpr.Args)
		if err != nil {
			return nil, true, fmt.Errorf("Upsert expects string field args: %w", err)
		}
		query.Upsert = true
		query.UpsertFields = fields
	case "Ignore":
		if len(callExpr.Args) != 0 {
			return nil, true, fmt.Errorf("Ignore does not accept arguments")
		}
		query.UpsertIgnore = true
	case "Name":
		if len(callExpr.Args) != 1 {
			return nil, true, fmt.Errorf("Name expects exactly one string argument")
		}
		name, err := parseSingleStringArg(callExpr.Args[0])
		if err != nil {
			return nil, true, fmt.Errorf("Name expects exactly one string argument: %w", err)
		}
		if !token.IsIdentifier(name) {
			return nil, true, fmt.Errorf("Name %q is not a valid identifier", name)
		}
		if suffix := util.ReservedNameSuffix(name); suffix != "" {
			return nil, true, fmt.Errorf("Name %q cannot end with %s, the generator appends it", name, suffix)
		}
		query.Name = name
	case "Contracts":
		contracts, err := parseQueryContracts(callExpr.Args)
		if err != nil {
			return nil, true, err
		}
		query.Contracts = contracts
	default:
		return nil, false, nil
	}

	return []schema.Query{query}, true, nil
}

func parseQueryContracts(args []ast.Expr) ([]schema.Contract, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("Contracts expects at least one of entlite.SQLC() or entlite.PROTO()")
	}

	var contracts []schema.Contract
	for _, arg := range args {
		callExpr, ok := arg.(*ast.CallExpr)
		if !ok {
			return nil, fmt.Errorf("Contracts expects entlite.SQLC() or entlite.PROTO()")
		}

		contract, err := parseContractCall(callExpr)
		if err != nil {
			return nil, err
		}
		if contract.Type == "" {
			return nil, fmt.Errorf("Contracts expects entlite.SQLC() or entlite.PROTO()")
		}
		if contract.Access != schema.AccessFull {
			return nil, fmt.Errorf("query contract %s cannot use ReadOnly or WriteOnly, a query is already a read or a write", contract.Type)
		}

		contracts = append(contracts, contract)
	}

	return contracts, nil
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

// parseColumnArg reads the single field name of an Asc()/Desc() call
func parseColumnArg(args []ast.Expr, method string) (string, error) {
	if len(args) != 1 {
		return "", fmt.Errorf("%s expects exactly one string field", method)
	}
	field, err := parseSingleStringArg(args[0])
	if err != nil {
		return "", fmt.Errorf("%s expects exactly one string field: %w", method, err)
	}
	return field, nil
}

func parseSingleStringArg(arg ast.Expr) (string, error) {
	lit, ok := arg.(*ast.BasicLit)
	if !ok || lit.Kind != token.STRING {
		return "", fmt.Errorf("expected string literal")
	}

	return strings.Trim(lit.Value, "\""), nil
}

func parseSingleIntArg(arg ast.Expr) (int, error) {
	lit, ok := arg.(*ast.BasicLit)
	if !ok || lit.Kind != token.INT {
		return 0, fmt.Errorf("expected int literal")
	}

	value, err := strconv.Atoi(lit.Value)
	if err != nil {
		return 0, fmt.Errorf("expected int literal: %w", err)
	}

	return value, nil
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

func applyQueryContracts(entity schema.Entity) ([]schema.Query, error) {
	queries := make([]schema.Query, 0, len(entity.Queries))

	for _, query := range entity.Queries {
		if len(query.Contracts) == 0 {
			for _, contract := range entity.Contracts {
				query.Contracts = append(query.Contracts, schema.Contract{Type: contract.Type})
			}
			queries = append(queries, query)
			continue
		}

		for _, contract := range query.Contracts {
			entityContract, ok := entity.GetContract(contract.Type)
			if !ok {
				return nil, fmt.Errorf("entity %q query %q declares contract %q, which the entity does not have", entity.Name, query.Type, contract.Type)
			}
			if entityContract.Access == schema.AccessRead && query.IsWrite() {
				return nil, fmt.Errorf("entity %q query %q cannot use contract %q, that entity contract is read only", entity.Name, query.Type, contract.Type)
			}
			if entityContract.Access == schema.AccessWrite && !query.IsWrite() {
				return nil, fmt.Errorf("entity %q query %q cannot use contract %q, that entity contract is write only", entity.Name, query.Type, contract.Type)
			}
		}

		queries = append(queries, query)
	}

	return queries, nil
}

func validateQueryFields(entity schema.Entity) error {
	if len(entity.Queries) == 0 {
		return nil
	}

	for _, query := range entity.Queries {
		if query.HasOffset && !query.HasLimit {
			return fmt.Errorf("entity %q query %q has Offset() without Limit()", entity.Name, query.Type)
		}

		if query.UpsertIgnore && !query.Upsert {
			return fmt.Errorf("entity %q query %q has Ignore() without Upsert()", entity.Name, query.Type)
		}

		if query.Upsert {
			if err := validateUpsertTarget(entity, query); err != nil {
				return err
			}
		}

		if err := validateDistinctColumns(entity, query); err != nil {
			return err
		}

		if err := validateOrderColumns(entity, query); err != nil {
			return err
		}

		switch query.Type {
		case schema.QueryGetBy:
			if len(query.Fields) == 0 {
				return fmt.Errorf("entity %q has query %q with empty fields", entity.Name, query.Type)
			}

			for _, fieldName := range query.Fields {
				if !entityHasField(entity, fieldName) {
					return fmt.Errorf("entity %q query %q references nonexisting field %q", entity.Name, query.Type, fieldName)
				}
				if entityFieldIsVirtual(entity, fieldName) {
					return fmt.Errorf("entity %q query %q references virtual field %q, which has no database column", entity.Name, query.Type, fieldName)
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
				if entityFieldIsVirtual(entity, fieldName) {
					return fmt.Errorf("entity %q query %q references virtual field %q, which has no database column", entity.Name, query.Type, fieldName)
				}
			}

			for _, queryFilter := range query.Filters {
				if !entityHasField(entity, queryFilter.Field) {
					return fmt.Errorf("entity %q query %q filter references nonexisting field %q", entity.Name, query.Type, queryFilter.Field)
				}
				if entityFieldIsVirtual(entity, queryFilter.Field) {
					return fmt.Errorf("entity %q query %q filter references virtual field %q, which has no database column", entity.Name, query.Type, queryFilter.Field)
				}
			}
		}
	}

	return nil
}

func validateDistinctColumns(entity schema.Entity, query schema.Query) error {
	if !query.HasDistinct() {
		return nil
	}

	if query.Count {
		return fmt.Errorf("entity %q query %q has Distinct() with Count(), the total would count the rows the dedupe drops", entity.Name, query.Type)
	}

	if key, found := entity.ContainsUniqueKey(query.Distinct); found {
		return fmt.Errorf("entity %q query %q Distinct selects the unique key (%s), which every row has a different value of, so nothing is deduplicated", entity.Name, query.Type, strings.Join(key, ", "))
	}

	seen := make(map[string]bool, len(query.Distinct))
	for _, fieldName := range query.Distinct {
		field, found := entity.GetFieldByName(fieldName)
		if !found {
			return fmt.Errorf("entity %q query %q Distinct references nonexisting field %q", entity.Name, query.Type, fieldName)
		}
		if entity.IsFieldVirtual(field) {
			return fmt.Errorf("entity %q query %q Distinct references virtual field %q, which has no database column", entity.Name, query.Type, fieldName)
		}
		if query.HasContract(schema.ContractPROTO) && !field.CanApiRead() {
			return fmt.Errorf("entity %q query %q Distinct returns field %q, which the proto contract cannot read", entity.Name, query.Type, fieldName)
		}
		lower := strings.ToLower(fieldName)
		if seen[lower] {
			return fmt.Errorf("entity %q query %q Distinct repeats field %q", entity.Name, query.Type, fieldName)
		}
		seen[lower] = true
	}

	return nil
}

// validateOrderColumns checks every Asc()/Desc() column is a real column, named once
func validateOrderColumns(entity schema.Entity, query schema.Query) error {
	seen := make(map[string]bool, len(query.OrderBy))
	for _, column := range query.OrderBy {
		if !entityHasField(entity, column.Name) {
			return fmt.Errorf("entity %q query %q order by references nonexisting field %q", entity.Name, query.Type, column.Name)
		}
		if entityFieldIsVirtual(entity, column.Name) {
			return fmt.Errorf("entity %q query %q order by references virtual field %q, which has no database column", entity.Name, query.Type, column.Name)
		}
		// a distinct select only holds its own columns, so nothing else can be sorted
		if query.HasDistinct() && !containsFold(query.Distinct, column.Name) {
			return fmt.Errorf("entity %q query %q sorts by field %q, which Distinct() does not select", entity.Name, query.Type, column.Name)
		}
		lower := strings.ToLower(column.Name)
		if seen[lower] {
			return fmt.Errorf("entity %q query %q order by repeats field %q", entity.Name, query.Type, column.Name)
		}
		seen[lower] = true
	}

	return nil
}

func containsFold(values []string, name string) bool {
	for _, value := range values {
		if strings.EqualFold(value, name) {
			return true
		}
	}

	return false
}

// validateUpsertTarget checks the conflict target is a key the database can collide
// on. A target that is not unique is only rejected when the query runs.
func validateUpsertTarget(entity schema.Entity, query schema.Query) error {
	seen := make(map[string]bool, len(query.UpsertFields))
	for _, fieldName := range query.UpsertFields {
		if !entityHasField(entity, fieldName) {
			return fmt.Errorf("entity %q query %q Upsert references nonexisting field %q", entity.Name, query.Type, fieldName)
		}
		if entityFieldIsVirtual(entity, fieldName) {
			return fmt.Errorf("entity %q query %q Upsert references virtual field %q, which has no database column", entity.Name, query.Type, fieldName)
		}
		lower := strings.ToLower(fieldName)
		if seen[lower] {
			return fmt.Errorf("entity %q query %q Upsert repeats field %q", entity.Name, query.Type, fieldName)
		}
		seen[lower] = true
	}

	// an insert leaves out a database generated key, so it never collides on one
	if len(query.UpsertFields) == 0 && entity.HasAutoGeneratedKey() {
		return fmt.Errorf("entity %q query %q has Upsert() without fields, but the primary key is generated by the database and never collides, name a unique field instead", entity.Name, query.Type)
	}

	target := entity.UpsertTarget(query)
	if len(target) == 0 {
		return fmt.Errorf("entity %q query %q has Upsert() without fields, but the entity has no primary key, name a unique field instead", entity.Name, query.Type)
	}

	if !entity.HasUniqueConstraint(target) {
		return fmt.Errorf("entity %q query %q Upsert target (%s) is not unique, it must be the primary key, a Unique() field or the fields of a Unique() index", entity.Name, query.Type, strings.Join(target, ", "))
	}

	// DO UPDATE needs a column to write, DO NOTHING does not
	if !query.UpsertIgnore && len(entity.UpsertSetFields(target)) == 0 {
		return fmt.Errorf("entity %q query %q has Upsert() with no column left to update, every other column is a key or Immutable, use Ignore() instead", entity.Name, query.Type)
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

func entityFieldIsVirtual(entity schema.Entity, fieldName string) bool {
	for _, field := range entity.Fields {
		if strings.EqualFold(field.Name, fieldName) {
			return entity.IsFieldVirtual(field)
		}
	}

	return false
}

func entityFieldIsOptional(entity schema.Entity, fieldName string) bool {
	for _, field := range entity.Fields {
		if strings.EqualFold(field.Name, fieldName) {
			return field.Optional
		}
	}

	return false
}

func entityFieldHasType(entity schema.Entity, fieldName string, fieldType schema.FieldType) bool {
	for _, field := range entity.Fields {
		if strings.EqualFold(field.Name, fieldName) {
			return field.Type == fieldType
		}
	}

	return false
}
