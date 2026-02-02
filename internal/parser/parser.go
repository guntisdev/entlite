package parser

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"

	"github.com/guntisdev/entlite/internal/schema"
)

type DiscoveredEntity struct {
	Name string
	Path string
}

func DiscoverEntities(entityDir string) ([]DiscoveredEntity, error) {
	var matches []string
	err := filepath.WalkDir(entityDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && filepath.Ext(path) == ".go" {
			matches = append(matches, path)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to find go files in %s: %w", entityDir, err)
	}

	fmt.Printf("Found %d files\n", len(matches))

	var entityList []DiscoveredEntity

	for _, file := range matches {

		entities, err := findEntitiesInFile(file)
		if err != nil {
			return nil, fmt.Errorf("failed to examine file %s: %w", file, err)
		}

		for _, entity := range entities {
			entityList = append(entityList, DiscoveredEntity{file, entity})
		}

		// TODO remove
		for _, entityName := range entities {
			fmt.Printf("Found entities %s\n", entityName)
		}
	}

	if len(entityList) == 0 {
		return nil, fmt.Errorf("no entities found in directory")
	}

	return entityList, nil
}

func findEntitiesInFile(filename string) ([]string, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filename, nil, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to parse file %s: %w", filename, err)
	}

	var entities []string

	for _, decl := range file.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.TYPE {
			continue
		}

		for _, spec := range genDecl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}

			structType, ok := typeSpec.Type.(*ast.StructType)
			if !ok {
				continue
			}

			if embedsEntliteSchema(structType) {
				entities = append(entities, typeSpec.Name.Name)
			}
		}
	}

	return entities, nil
}

func embedsEntliteSchema(structType *ast.StructType) bool {
	if structType.Fields == nil {
		return false
	}

	for _, field := range structType.Fields.List {
		if len(field.Names) == 0 {
			switch fieldType := field.Type.(type) {
			case *ast.SelectorExpr:
				if ident, ok := fieldType.X.(*ast.Ident); ok && ident.Name == "entlite" {
					if fieldType.Sel.Name == "Schema" {
						return true
					}
				}
			}
		}
	}

	return false
}

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

		if funcDecl.Name.Name == "Annotations" {
			annotations, err := parseAnnotationsMethod(funcDecl)
			if err != nil {
				return entity, fmt.Errorf("failed to parse annotations: %w", err)
			}
			entity.Annotations = annotations
		}

		// parse Fields

	}

	return entity, nil
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

				// TODO delete
				fmt.Printf("Service annotation found with %d arguments\n", len(callExpr.Args))
				if len(callExpr.Args) > 0 {
					methods := parseServiceArguments(callExpr.Args)
					annotation.Methods = methods
					// TODO delete
					fmt.Printf("Parser methods form Service arguments: %v\n", methods)
				}
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
				if ident, ok := selExpr.X.(*ast.Ident); ok && ident.Name == "entlite" {
					if selExpr.Sel.Name == "Methods" {
						// TODO remove
						fmt.Printf("Found Method() call with %d arguments \n", len(callExpr.Args))
						if len(callExpr.Args) > 0 {
							methodMethods := parseMethodsArguments(callExpr.Args[0])
							methods = append(methods, methodMethods...)
							// TODO remove
							fmt.Printf("Extract methods: %v\n", methodMethods)
						}
					}
				}
			}
		}
	}

	return methods
}

func parseMethodsArguments(expr ast.Expr) []schema.Method {
	var methods []schema.Method

	if binExpr, ok := expr.(*ast.BinaryExpr); ok && binExpr.Op == token.OR {
		// Recursively parse both sides ofthe OR
		leftMethods := parseMethodsArguments(binExpr.X)
		rightMethods := parseMethodsArguments(binExpr.Y)
		methods = append(methods, leftMethods...)
		methods = append(methods, rightMethods...)
	} else if selExpr, ok := expr.(*ast.SelectorExpr); ok {
		if ident, ok := selExpr.X.(*ast.Ident); ok && ident.Name == "entlite" {
			switch selExpr.Sel.Name {
			case "MethodCreate":
				methods = append(methods, schema.MethodCreate)
			case "MethodGet":
				methods = append(methods, schema.MethodGet)
			case "MethodUpdate":
				methods = append(methods, schema.MethodUpdate)
			case "MethodDelete":
				methods = append(methods, schema.MethodUpdate)
			case "MethodList":
				methods = append(methods, schema.MethodList)
			}
			// TODO
			fmt.Printf("Found methods: %s\n", selExpr.Sel.Name)
		}
	}

	return methods
}
