package docs

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

type apiPackage struct {
	Name   string
	Doc    string
	Types  []apiType
	Funcs  []apiFunc
	Consts []apiConst
}

type apiType struct {
	Name        string
	Doc         string
	IsInterface bool
	Methods     []apiMethod
}

type apiMethod struct {
	Name      string
	Signature string // e.g. Unique() StringFieldBuilder
	Doc       string
}

type apiFunc struct {
	Name      string
	Signature string // e.g. String(name string) StringFieldBuilder
	Result    string // the single named result type, empty when there is none
	Doc       string
}

type apiConst struct {
	Name  string
	Value string
	Doc   string
}

func (p apiPackage) Type(name string) (apiType, bool) {
	for _, t := range p.Types {
		if t.Name == name {
			return t, true
		}
	}

	return apiType{}, false
}

func (p apiPackage) FuncsReturning(suffix string) []apiFunc {
	var found []apiFunc
	for _, f := range p.Funcs {
		if strings.HasSuffix(f.Result, suffix) {
			found = append(found, f)
		}
	}

	return found
}

func (t apiType) MethodNames() []string {
	names := make([]string, len(t.Methods))
	for i, m := range t.Methods {
		names[i] = m.Name
	}

	return names
}

func (t apiType) HasMethod(name string) bool {
	for _, m := range t.Methods {
		if m.Name == name {
			return true
		}
	}

	return false
}

func parseAPI(dir string) (apiPackage, error) {
	files, fset, err := parseGoFiles(dir)
	if err != nil {
		return apiPackage{}, err
	}

	pkg := apiPackage{Name: "", Doc: "", Types: nil, Funcs: nil, Consts: nil}
	typeIndex := map[string]int{}

	// first pass: package doc, types and constants
	for _, file := range files {
		pkg.Name = file.Name.Name
		if file.Doc != nil && pkg.Doc == "" {
			pkg.Doc = docLine("Package "+pkg.Name, file.Doc.Text())
		}

		for _, decl := range file.Decls {
			genDecl, ok := decl.(*ast.GenDecl)
			if !ok {
				continue
			}

			switch genDecl.Tok {
			case token.TYPE:
				for _, spec := range genDecl.Specs {
					typeSpec, ok := spec.(*ast.TypeSpec)
					if !ok || !typeSpec.Name.IsExported() {
						continue
					}
					typeIndex[typeSpec.Name.Name] = len(pkg.Types)
					pkg.Types = append(pkg.Types, newAPIType(fset, typeSpec, genDecl))
				}
			case token.CONST:
				pkg.Consts = append(pkg.Consts, newAPIConsts(fset, genDecl)...)
			}
		}
	}

	// second pass: functions and the methods of the types above
	for _, file := range files {
		for _, decl := range file.Decls {
			funcDecl, ok := decl.(*ast.FuncDecl)
			if !ok || !funcDecl.Name.IsExported() {
				continue
			}
			if isMarker(funcDecl.Name.Name, funcDecl.Type) {
				continue
			}

			if funcDecl.Recv == nil {
				pkg.Funcs = append(pkg.Funcs, newAPIFunc(fset, funcDecl))
				continue
			}

			idx, ok := typeIndex[receiverName(funcDecl.Recv)]
			if !ok {
				continue
			}
			pkg.Types[idx].Methods = append(pkg.Types[idx].Methods, apiMethod{
				Name:      funcDecl.Name.Name,
				Signature: signature(fset, funcDecl.Name.Name, funcDecl.Type),
				Doc:       docLine(funcDecl.Name.Name, funcDecl.Doc.Text()),
			})
		}
	}

	return pkg, nil
}

// parses every non test go file of a directory, in file name order
func parseGoFiles(dir string) ([]*ast.File, *token.FileSet, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read %s: %w", dir, err)
	}

	fset := token.NewFileSet()
	var files []*ast.File

	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}

		file, err := parser.ParseFile(fset, filepath.Join(dir, name), nil, parser.ParseComments)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to parse %s: %w", name, err)
		}
		files = append(files, file)
	}

	return files, fset, nil
}

func newAPIType(fset *token.FileSet, typeSpec *ast.TypeSpec, genDecl *ast.GenDecl) apiType {
	doc := typeSpec.Doc.Text()
	if doc == "" {
		doc = genDecl.Doc.Text()
	}

	result := apiType{
		Name:        typeSpec.Name.Name,
		Doc:         docLine(typeSpec.Name.Name, doc),
		IsInterface: false,
		Methods:     nil,
	}

	iface, ok := typeSpec.Type.(*ast.InterfaceType)
	if !ok {
		return result
	}
	result.IsInterface = true

	for _, field := range iface.Methods.List {
		funcType, ok := field.Type.(*ast.FuncType)
		if !ok || len(field.Names) == 0 {
			continue // an embedded interface, its own page section covers it
		}

		name := field.Names[0].Name
		if !field.Names[0].IsExported() || isMarker(name, funcType) {
			continue
		}
		result.Methods = append(result.Methods, apiMethod{
			Name:      name,
			Signature: signature(fset, name, funcType),
			Doc:       docLine(name, field.Doc.Text()),
		})
	}

	return result
}

func newAPIConsts(fset *token.FileSet, genDecl *ast.GenDecl) []apiConst {
	var consts []apiConst

	for _, spec := range genDecl.Specs {
		valueSpec, ok := spec.(*ast.ValueSpec)
		if !ok || len(valueSpec.Names) == 0 || !valueSpec.Names[0].IsExported() {
			continue
		}

		value := ""
		if len(valueSpec.Values) > 0 {
			value = strings.Trim(exprString(fset, valueSpec.Values[0]), `"`)
		}

		consts = append(consts, apiConst{
			Name:  valueSpec.Names[0].Name,
			Value: value,
			Doc:   docLine(valueSpec.Names[0].Name, valueSpec.Doc.Text()),
		})
	}

	return consts
}

// newAPIFunc reads a package level function.
func newAPIFunc(fset *token.FileSet, funcDecl *ast.FuncDecl) apiFunc {
	result := ""
	if funcDecl.Type.Results != nil && len(funcDecl.Type.Results.List) == 1 {
		result = exprString(fset, funcDecl.Type.Results.List[0].Type)
	}

	return apiFunc{
		Name:      funcDecl.Name.Name,
		Signature: signature(fset, funcDecl.Name.Name, funcDecl.Type),
		Result:    result,
		Doc:       docLine(funcDecl.Name.Name, funcDecl.Doc.Text()),
	}
}

func isMarker(name string, funcType *ast.FuncType) bool {
	if funcType.Params != nil && len(funcType.Params.List) > 0 {
		return false
	}
	if funcType.Results != nil && len(funcType.Results.List) > 0 {
		return false
	}

	switch name {
	case "Field", "Query", "Index", "Filter", "Contract":
		return true
	}

	return false
}

func receiverName(recv *ast.FieldList) string {
	if recv == nil || len(recv.List) == 0 {
		return ""
	}

	switch t := recv.List[0].Type.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		if ident, ok := t.X.(*ast.Ident); ok {
			return ident.Name
		}
	}

	return ""
}

func signature(fset *token.FileSet, name string, funcType *ast.FuncType) string {
	return name + strings.TrimPrefix(exprString(fset, funcType), "func")
}

func exprString(fset *token.FileSet, node ast.Node) string {
	var buf bytes.Buffer
	if err := printer.Fprint(&buf, fset, node); err != nil {
		return ""
	}

	return buf.String()
}

func docLine(name, doc string) string {
	text := strings.TrimSpace(doc)
	if text == "" {
		return ""
	}

	// keep the first paragraph only, on a single line
	text = strings.Join(strings.Fields(strings.Split(text, "\n\n")[0]), " ")
	text = strings.TrimSpace(strings.TrimPrefix(text, name))
	// "Column is a single column" reads better as "A single column"
	for _, verb := range []string{"is ", "are "} {
		text = strings.TrimPrefix(text, verb)
	}

	runes := []rune(text)
	if len(runes) == 0 {
		return ""
	}
	runes[0] = unicode.ToUpper(runes[0])

	return string(runes)
}
