package parser

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"

	"github.com/guntisdev/entlite/internal/schema"
)

func ParseEntities(discoveredEntities []DiscoveredEntity) ([]schema.Entity, error) {
	var entities []schema.Entity

	for _, discovered := range discoveredEntities {
		parsed, err := parseEntityFromFile(discovered)
		if err != nil {
			continue
		}
		entities = append(entities, parsed)
	}

	return entities, nil
}

func parseEntityFromFile(discovered DiscoveredEntity) (schema.Entity, error) {
	entity := schema.Entity{
		Name: discovered.Name,
	}
	// TODO remove
	fmt.Printf("parseFromFile %s\n", discovered.Name)

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, discovered.Path, nil, parser.ParseComments)
	if err != nil {
		return entity, fmt.Errorf("failed to parse file %s: %w", discovered.Path, err)
	}

	for _, decl := range file.Decls {
		funcDecl, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}

		if funcDecl.Recv == nil || len(funcDecl.Recv.List) == 0 {
			continue
		}

		recvType := funcDecl.Recv.List[0].Type
		var recvTypeName string

		switch t := recvType.(type) {
		case *ast.Ident:
			recvTypeName = t.Name
		case *ast.StarExpr:
			if ident, ok := t.X.(*ast.Ident); ok {
				recvTypeName = ident.Name
			}
		}

		if recvTypeName != entity.Name {
			continue
		}

		// Parse Annotations
		if funcDecl.Name.Name == "Annotations" {
			annotations, err := parseAnnotationsMethod(funcDecl)
			if err != nil {
				return entity, fmt.Errorf("failed to parse annotations: %w", err)
			}
			entity.Annotations = annotations
		}

		// Parse Fields
		if funcDecl.Name.Name == "Fields" {
			fields, err := parseFieldsMethod(funcDecl)
			if err != nil {
				return entity, fmt.Errorf("failed to parse fields: %w", err)
			}

			if err := checkProtoFieldCollision(fields); err != nil {
				return entity, err
			}

			// add protoField, add id if not there
			fields = addFieldNumbers(fields)
			entity.Fields = fields
		}

	}

	return entity, nil
}

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
					field := parseFieldExpression(elt)
					if field.Name != "" {
						fields = append(fields, field)
					}
				}

			}
		}
	}

	return fields, nil
}

func parseFieldExpression(expr ast.Expr) schema.Field {
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
							field.Name = strings.Trim(lit.Value, "\"")
						}
					}
				case "Int32":
					field.Type = schema.FieldTypeInt32
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
				case "DefaultNow":
					field.DefaultNow = true
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

	return field
}

func parseDefaultValue(expr ast.Expr) any {
	switch e := expr.(type) {
	case *ast.BasicLit:
		switch e.Kind {
		case token.STRING:
			return strings.Trim(e.Value, "\"")
		case token.INT:
			var val int
			fmt.Scanf(e.Value, "%d", &val)
			return val
		case token.FLOAT:
			var val float32
			fmt.Scanf(e.Value, "%f", &val)
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
	if _, err := fmt.Scanf(s, "%d", &i); err == nil {
		return &i
	}
	return nil
}

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
			case "Service":
				annotation.Type = schema.AnnotationService

				var methods []schema.Method
				if len(callExpr.Args) > 0 {
					methods = parseServiceArguments(callExpr.Args)
				} else {
					// if no arguments in Service() then all default methods
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
