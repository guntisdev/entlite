package parser

import (
	"fmt"
	"go/ast"
	"go/token"
	"strings"

	"github.com/guntisdev/entlite/internal/schema"
	"github.com/guntisdev/entlite/pkg/entlite/permissions"
)

func parseFieldsMethod(funcDecl *ast.FuncDecl) ([]schema.Field, error) {
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
				for _, elt := range compLit.Elts {
					field, err := parseFieldExpression(elt)
					if err != nil {
						return nil, err
					}
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

	field.Permissions = permissions.Default // default all permission

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
							field.Name = strings.Trim(lit.Value, "\"")
						}
					}
				case "Int":
					field.Type = schema.FieldTypeInt
					if len(e.Args) > 0 {
						if lit, ok := e.Args[0].(*ast.BasicLit); ok && lit.Kind == token.STRING {
							field.Name = strings.Trim(lit.Value, "\"")
						}
					}
				case "Int64":
					field.Type = schema.FieldTypeInt64
					if len(e.Args) > 0 {
						if lit, ok := e.Args[0].(*ast.BasicLit); ok && lit.Kind == token.STRING {
							field.Name = strings.Trim(lit.Value, "\"")
						}
					}
				case "Float":
					field.Type = schema.FieldTypeFloat
					if len(e.Args) > 0 {
						if lit, ok := e.Args[0].(*ast.BasicLit); ok && lit.Kind == token.STRING {
							field.Name = strings.Trim(lit.Value, "\"")
						}
					}
				case "Bool":
					field.Type = schema.FieldTypeBool
					if len(e.Args) > 0 {
						if lit, ok := e.Args[0].(*ast.BasicLit); ok && lit.Kind == token.STRING {
							field.Name = strings.Trim(lit.Value, "\"")
						}
					}
				case "Time":
					field.Type = schema.FieldTypeTime
					if len(e.Args) > 0 {
						if lit, ok := e.Args[0].(*ast.BasicLit); ok && lit.Kind == token.STRING {
							field.Name = strings.Trim(lit.Value, "\"")
						}
					}
				case "Byte":
					field.Type = schema.FieldTypeByte
					if len(e.Args) > 0 {
						if lit, ok := e.Args[0].(*ast.BasicLit); ok && lit.Kind == token.STRING {
							field.Name = strings.Trim(lit.Value, "\"")
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
				case "Comment":
					if len(e.Args) > 0 {
						if lit, ok := e.Args[0].(*ast.BasicLit); ok && lit.Kind == token.STRING {
							field.Comment = strings.Trim(lit.Value, "\"")
						}
					}
				case "Permissions":
					if len(e.Args) > 0 {
						field.Permissions = parsePermissionsExpression(e.Args[0])
					}
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

func parseDefaultValue(expr ast.Expr) any {
	switch e := expr.(type) {
	case *ast.BasicLit:
		switch e.Kind {
		case token.STRING:
			return strings.Trim(e.Value, "\"")
		case token.INT:
			var val int
			fmt.Sscanf(e.Value, "%d", &val)
			return val
		case token.FLOAT:
			var val float32
			fmt.Sscanf(e.Value, "%f", &val)
			return val
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

func parsePermissionsExpression(expr ast.Expr) permissions.Permission {
	var perm permissions.Permission

	if binExpr, ok := expr.(*ast.BinaryExpr); ok && binExpr.Op == token.OR {
		// Handle binary OR expressions like permissions.DbRead | permissions.ApiRead
		leftPerm := parsePermissionsExpression(binExpr.X)
		rightPerm := parsePermissionsExpression(binExpr.Y)
		perm = leftPerm | rightPerm
	} else if selExpr, ok := expr.(*ast.SelectorExpr); ok {
		// Handle selector expressions like permissions.Standard
		if ident, ok := selExpr.X.(*ast.Ident); ok && ident.Name == "permissions" {
			switch selExpr.Sel.Name {
			case "DbRead":
				perm = permissions.DbRead
			case "DbWrite":
				perm = permissions.DbWrite
			case "ApiRead":
				perm = permissions.ApiRead
			case "ApiWrite":
				perm = permissions.ApiWrite
			case "Default":
				perm = permissions.Default
			case "ReadOnly":
				perm = permissions.ReadOnly
			case "WriteOnly":
				perm = permissions.WriteOnly
			case "Internal":
				perm = permissions.Internal
			case "Virtual":
				perm = permissions.Virtual
			}
		}
	}

	return perm
}
