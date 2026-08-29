package parser

import (
	"fmt"
	"go/ast"
	"go/token"
	"strconv"
	"strings"

	"github.com/guntisdev/entlite/internal/schema"
)

func parseFieldsMethod(funcDecl *ast.FuncDecl, comments commentLookup) ([]schema.Field, error) {
	var fields []schema.Field

	if funcDecl.Body == nil {
		return fields, nil
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
					field, err := parseFieldExpression(elt)
					if err != nil {
						return nil, err
					}
					field.Comment = comments.docAbove(elt.Pos(), prevEnd)
					prevEnd = elt.End()
					if field.Name != "" {
						fields = append(fields, field)
					}
				}

			}
		}
	}

	return fields, nil
}

func parseFieldExpression(expr ast.Expr) (schema.Field, error) {
	field := schema.Field{}

	// Handle method chaining like entlite.String("name").ProtoField(2)
	currentExpr := expr

	for currentExpr != nil {
		switch e := currentExpr.(type) {
		case *ast.CallExpr:
			if selExpr, ok := e.Fun.(*ast.SelectorExpr); ok {
				methodName := selExpr.Sel.Name

				switch methodName {
				case "String":
					field.Type = schema.FieldTypeString
					if len(e.Args) > 0 {
						if lit, ok := e.Args[0].(*ast.BasicLit); ok && lit.Kind == token.STRING {
							field.Name = unquote(lit.Value)
						}
					}
				case "Int":
					field.Type = schema.FieldTypeInt
					if len(e.Args) > 0 {
						if lit, ok := e.Args[0].(*ast.BasicLit); ok && lit.Kind == token.STRING {
							field.Name = unquote(lit.Value)
						}
					}
				case "Int64":
					field.Type = schema.FieldTypeInt64
					if len(e.Args) > 0 {
						if lit, ok := e.Args[0].(*ast.BasicLit); ok && lit.Kind == token.STRING {
							field.Name = unquote(lit.Value)
						}
					}
				case "Float":
					field.Type = schema.FieldTypeFloat
					if len(e.Args) > 0 {
						if lit, ok := e.Args[0].(*ast.BasicLit); ok && lit.Kind == token.STRING {
							field.Name = unquote(lit.Value)
						}
					}
				case "Bool":
					field.Type = schema.FieldTypeBool
					if len(e.Args) > 0 {
						if lit, ok := e.Args[0].(*ast.BasicLit); ok && lit.Kind == token.STRING {
							field.Name = unquote(lit.Value)
						}
					}
				case "Time":
					field.Type = schema.FieldTypeTime
					if len(e.Args) > 0 {
						if lit, ok := e.Args[0].(*ast.BasicLit); ok && lit.Kind == token.STRING {
							field.Name = unquote(lit.Value)
						}
					}
				case "Byte":
					field.Type = schema.FieldTypeByte
					if len(e.Args) > 0 {
						if lit, ok := e.Args[0].(*ast.BasicLit); ok && lit.Kind == token.STRING {
							field.Name = unquote(lit.Value)
						}
					}
				case "JSON":
					field.Type = schema.FieldTypeJSON
					if len(e.Args) > 0 {
						if lit, ok := e.Args[0].(*ast.BasicLit); ok && lit.Kind == token.STRING {
							field.Name = unquote(lit.Value)
						}
					}
				case "ProtoField":
					if len(e.Args) > 0 {
						if lit, ok := e.Args[0].(*ast.BasicLit); ok && lit.Kind == token.INT {
							if val := parseInt(lit.Value); val != nil {
								field.ProtoField = *val
							}
						}
					}
				case "Contracts":
					contracts, err := parseFieldContracts(e.Args)
					if err != nil {
						return field, err
					}
					field.Contracts = contracts
				case "Unique":
					field.Unique = true
				case "Immutable":
					field.Immutable = true
				case "Optional":
					field.Optional = true
				case "Default":
					if len(e.Args) > 0 {
						field.DefaultValue = parseDefaultValue(e.Args[0])
					}
				case "DefaultFunc":
					if len(e.Args) > 0 {
						fn, err := parseDefaultFuncValue(e.Args[0])
						if err != nil {
							return field, fmt.Errorf("field %q: %w", field.Name, err)
						}
						field.DefaultFunc = fn
					}
				case "Validate":
					if len(e.Args) > 0 {
						fn, err := parseValidateFuncValue(e.Args[0])
						if err != nil {
							return field, fmt.Errorf("field %q: %w", field.Name, err)
						}
						field.Validate = fn
					}
				}

				// Continue with the receiver of this method call
				currentExpr = selExpr.X

			} else {
				// not a method call, check if it is a top level function call
				if _, ok := e.Fun.(*ast.Ident); ok {
					// Handle calls like entlite.String
					break
				}
				currentExpr = nil
			}
		default:
			currentExpr = nil
		}
	}

	return field, nil
}

func unquote(raw string) string {
	if val, err := strconv.Unquote(raw); err == nil {
		return val
	}
	return strings.Trim(raw, "\"")
}

func parseDefaultValue(expr ast.Expr) any {
	switch e := expr.(type) {
	case *ast.BasicLit:
		switch e.Kind {
		case token.STRING:
			return unquote(e.Value)
		case token.INT:
			// base 0 also handles 0x, 0b, 0o and underscores
			val, err := strconv.ParseInt(e.Value, 0, 64)
			if err != nil {
				return nil
			}
			return int(val)
		case token.FLOAT:
			val, err := strconv.ParseFloat(e.Value, 32)
			if err != nil {
				return nil
			}
			return float32(val)
		}
	case *ast.Ident:
		if e.Name == "true" {
			return true
		}
		if e.Name == "false" {
			return false
		}
	}
	return nil
}

func parseInt(s string) *int {
	var i int
	if _, err := fmt.Sscanf(s, "%d", &i); err == nil {
		return &i
	}
	return nil
}

func parseDefaultFuncValue(expr ast.Expr) (func() any, error) {
	switch e := expr.(type) {
	case *ast.SelectorExpr:
		// Accept package function references like uuid.NewString, time.Now
		if ident, ok := e.X.(*ast.Ident); ok {
			pkg := ident.Name
			fn := e.Sel.Name
			return func() any {
				// Placeholder - stores function reference as pkg.Function
				return fmt.Sprintf("%s.%s", pkg, fn)
			}, nil
		}
	case *ast.Ident:
		// Accept direct function references like someFunction
		fnName := e.Name
		return func() any {
			return fnName
		}, nil
	case *ast.FuncLit:
		return nil, fmt.Errorf("default func cannot be an anonymous function, use a named function reference instead")
	}
	return nil, fmt.Errorf("default func must be a function reference")
}

func parseValidateFuncValue(expr ast.Expr) (func() any, error) {
	switch e := expr.(type) {
	case *ast.SelectorExpr:
		// Accept package function references like validators.CheckEmail
		if ident, ok := e.X.(*ast.Ident); ok {
			pkg := ident.Name
			fn := e.Sel.Name
			return func() any {
				// Placeholder - stores function reference as pkg.Function
				return fmt.Sprintf("%s.%s", pkg, fn)
			}, nil
		}
	case *ast.Ident:
		// Accept direct function references like MyValidateFunc
		fnName := e.Name
		return func() any {
			return fnName
		}, nil
	case *ast.FuncLit:
		return nil, fmt.Errorf("validate cannot be an anonymous function, use a named function reference instead")
	}
	return nil, fmt.Errorf("validate must be a function reference")
}

// applyFieldContracts fills in the entity contracts where a field declares none
// and checks the declared ones against the entity
func applyFieldContracts(entity schema.Entity) ([]schema.Field, error) {
	fields := make([]schema.Field, 0, len(entity.Fields))

	for _, field := range entity.Fields {
		if len(field.Contracts) == 0 {
			for _, contract := range entity.Contracts {
				field.Contracts = append(field.Contracts, schema.Contract{Type: contract.Type})
			}
			fields = append(fields, field)
			continue
		}

		for _, contract := range field.Contracts {
			if _, ok := entity.GetContract(contract.Type); !ok {
				return nil, fmt.Errorf("entity %q field %q declares contract %q, which the entity does not have", entity.Name, field.Name, contract.Type)
			}
		}

		fields = append(fields, field)
	}

	return fields, nil
}

// parseFieldContracts reads the contracts of a single field
func parseFieldContracts(args []ast.Expr) ([]schema.Contract, error) {
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

		contracts = append(contracts, contract)
	}

	return contracts, nil
}
