package sqlcwrap

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"github.com/guntisdev/entlite/internal/schema"
)

func Generate(inputFilePath string, parsedEntities []schema.Entity) (string, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, inputFilePath, nil, parser.ParseComments)
	if err != nil {
		return "", fmt.Errorf("failed to parse file: %w", err)
	}

	inputPackageName := node.Name.Name
	absInputDir, _ := filepath.Abs(filepath.Dir(inputFilePath))
	moduleName, workspaceRoot, err := findModuleInfo(absInputDir)
	if err != nil {
		return "", fmt.Errorf("failed to find module info: %w", err)
	}
	relPath, _ := filepath.Rel(workspaceRoot, absInputDir)
	importPath := filepath.Join(moduleName, relPath)
	importPath = filepath.ToSlash(importPath)

	entityMap := make(map[string]schema.Entity)
	for _, entity := range parsedEntities {
		entityMap[entity.Name] = entity
	}

	createParamsStructs := make(map[string]*ast.StructType)
	createFuncs := make(map[string]*ast.FuncDecl)

	for _, decl := range node.Decls {
		switch d := decl.(type) {
		case *ast.GenDecl:
			for _, spec := range d.Specs {
				if typeSpec, ok := spec.(*ast.TypeSpec); ok {
					if structType, ok := typeSpec.Type.(*ast.StructType); ok {
						// Check if this is a Create{Entity}Params struct
						if strings.HasPrefix(typeSpec.Name.Name, "Create") && strings.HasSuffix(typeSpec.Name.Name, "Params") {
							createParamsStructs[typeSpec.Name.Name] = structType
						}
					}
				}
			}
		case *ast.FuncDecl:
			// Find Create{Entity} functions
			if strings.HasPrefix(d.Name.Name, "Create") && d.Recv != nil {
				createFuncs[d.Name.Name] = d
			}
		}
	}

	var sb strings.Builder

	packageName := filepath.Base(filepath.Dir(absInputDir))
	sb.WriteString(fmt.Sprintf("package %s\n\n", packageName))

	needsContextImport := false
	needsTimeImport := false
	needsSQLImport := false

	// TODO find better way get imports (not hardcoded Time)
	for structName := range createParamsStructs {
		entityName := strings.TrimSuffix(strings.TrimPrefix(structName, "Create"), "Params")
		if entity, ok := entityMap[entityName]; ok {
			if hasDefaultFuncFields(entity) {
				needsContextImport = true
				for _, field := range entity.Fields {
					if field.DefaultFunc != nil && field.Type == schema.FieldTypeTime {
						needsTimeImport = true
					}
				}
			}
		}
		if structType, ok := createParamsStructs[structName]; ok {
			if usesSQLTypes(structType) {
				needsSQLImport = true
			}
		}
	}

	sb.WriteString("import (\n")
	if needsContextImport {
		sb.WriteString("\t\"context\"\n")
	}
	if needsSQLImport {
		sb.WriteString("\t\"database/sql\"\n")
	}
	if needsTimeImport {
		sb.WriteString("\t\"time\"\n")
	}
	sb.WriteString(fmt.Sprintf("\t%s \"%s\"\n", inputPackageName, importPath))
	sb.WriteString(")\n\n")

	for _, decl := range node.Decls {
		switch d := decl.(type) {
		case *ast.GenDecl:
			for _, spec := range d.Specs {
				switch s := spec.(type) {
				case *ast.TypeSpec:
					if s.Name.IsExported() {
						if strings.HasPrefix(s.Name.Name, "Create") && strings.HasSuffix(s.Name.Name, "Params") {
							entityName := strings.TrimSuffix(strings.TrimPrefix(s.Name.Name, "Create"), "Params")
							if entity, ok := entityMap[entityName]; ok && hasDefaultFuncFields(entity) {
								// Generate custom struct without DefaultFunc fields
								sb.WriteString(generateCreateStruct(s.Name.Name, createParamsStructs[s.Name.Name], entity, fset))
								continue
							}
						}
						// For Queries type, use a proper type (not alias) so we can add methods
						if s.Name.Name == "Queries" {
							sb.WriteString(fmt.Sprintf("type %s %s.%s\n", s.Name.Name, inputPackageName, s.Name.Name))
						} else {
							sb.WriteString(fmt.Sprintf("type %s = %s.%s\n", s.Name.Name, inputPackageName, s.Name.Name))
						}
					}
				case *ast.ValueSpec:
					for _, name := range s.Names {
						if name.IsExported() {
							if d.Tok == token.CONST {
								sb.WriteString(fmt.Sprintf("const %s = %s.%s\n", name.Name, inputPackageName, name.Name))
							} else {
								sb.WriteString(fmt.Sprintf("var %s = %s.%s\n", name.Name, inputPackageName, name.Name))
							}
						}
					}
				}
			}
		case *ast.FuncDecl:
			if d.Name.IsExported() && d.Recv == nil {
				sb.WriteString(fmt.Sprintf("var %s = %s.%s\n", d.Name.Name, inputPackageName, d.Name.Name))
			} else if d.Recv != nil && strings.HasPrefix(d.Name.Name, "Create") {
				entityName := strings.TrimPrefix(d.Name.Name, "Create")
				if entity, ok := entityMap[entityName]; ok && hasDefaultFuncFields(entity) {
					sb.WriteString(generateCreateMethodWrapper(d, entity, inputPackageName, fset))
					continue
				}
			}
		}
	}

	return sb.String(), nil
}

func hasDefaultFuncFields(entity schema.Entity) bool {
	for _, field := range entity.Fields {
		if field.DefaultFunc != nil {
			return true
		}
	}
	return false
}

func usesSQLTypes(structType *ast.StructType) bool {
	for _, field := range structType.Fields.List {
		if usesSQLType(field.Type) {
			return true
		}
	}
	return false
}

func usesSQLType(expr ast.Expr) bool {
	switch t := expr.(type) {
	case *ast.SelectorExpr:
		if ident, ok := t.X.(*ast.Ident); ok {
			return ident.Name == "sql"
		}
	case *ast.StarExpr:
		return usesSQLType(t.X)
	case *ast.ArrayType:
		return usesSQLType(t.Elt)
	}
	return false
}

func generateCreateStruct(structName string, structType *ast.StructType, entity schema.Entity, fset *token.FileSet) string {
	var sb strings.Builder

	defaultFuncFields := make(map[string]bool)
	for _, field := range entity.Fields {
		if field.DefaultFunc != nil {
			defaultFuncFields[toExportedName(field.Name)] = true
		}
	}

	sb.WriteString(fmt.Sprintf("type %s struct {\n", structName))

	for _, field := range structType.Fields.List {
		if len(field.Names) > 0 {
			fieldName := field.Names[0].Name
			// Skip fields that have DefaultFunc
			if !defaultFuncFields[fieldName] {
				sb.WriteString(fmt.Sprintf("\t%s %s", fieldName, formatType(field.Type)))
				if field.Tag != nil {
					sb.WriteString(fmt.Sprintf(" %s", field.Tag.Value))
				}
				sb.WriteString("\n")
			}
		}
	}

	sb.WriteString("}\n\n")
	return sb.String()
}

func generateCreateMethodWrapper(funcDecl *ast.FuncDecl, entity schema.Entity, inputPkg string, fset *token.FileSet) string {
	var sb strings.Builder

	receiverType := formatType(funcDecl.Recv.List[0].Type)
	sb.WriteString(fmt.Sprintf("func (q %s) %s(ctx context.Context, arg %sParams) ", receiverType, funcDecl.Name.Name, funcDecl.Name.Name))

	if funcDecl.Type.Results != nil && len(funcDecl.Type.Results.List) > 0 {
		sb.WriteString("(")
		for i, result := range funcDecl.Type.Results.List {
			if i > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(formatType(result.Type))
		}
		sb.WriteString(")")
	}

	sb.WriteString(" {\n")
	sb.WriteString(fmt.Sprintf("\tinternalArg := %s.%sParams{\n", inputPkg, funcDecl.Name.Name))

	defaultFuncFields := make(map[string]schema.Field)
	for _, field := range entity.Fields {
		if field.DefaultFunc != nil {
			defaultFuncFields[toExportedName(field.Name)] = field
		}
	}

	for _, field := range entity.Fields {
		exportedName := toExportedName(field.Name)
		if field.IsID() && field.DefaultFunc == nil {
			continue
		}
		if _, hasDefaultFunc := defaultFuncFields[exportedName]; hasDefaultFunc {
			funcName := field.DefaultFunc().(string)
			sb.WriteString(fmt.Sprintf("\t\t%s: %s(),\n", exportedName, funcName))
		} else {
			sb.WriteString(fmt.Sprintf("\t\t%s: arg.%s,\n", exportedName, exportedName))
		}
	}

	sb.WriteString("\t}\n")

	sb.WriteString(fmt.Sprintf("\treturn (*%s.Queries)(q).%s(ctx, internalArg)\n", inputPkg, funcDecl.Name.Name))
	sb.WriteString("}\n\n")

	return sb.String()
}

func formatType(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return "*" + formatType(t.X)
	case *ast.SelectorExpr:
		return formatType(t.X) + "." + t.Sel.Name
	case *ast.ArrayType:
		return "[]" + formatType(t.Elt)
	default:
		return "interface{}"
	}
}

func toExportedName(name string) string {
	parts := strings.Split(name, "_")
	for i := range parts {
		if len(parts[i]) > 0 {
			parts[i] = strings.ToUpper(parts[i][:1]) + parts[i][1:]
		}
	}
	return strings.Join(parts, "")
}

func findModuleInfo(startDir string) (string, string, error) {
	dir := startDir

	for {
		goModPath := filepath.Join(dir, "go.mod")
		if _, err := os.Stat(goModPath); err == nil {
			content, err := os.ReadFile(goModPath)
			if err != nil {
				return "", "", err
			}

			lines := strings.Split(string(content), "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if strings.HasPrefix(line, "module ") {
					moduleName := strings.TrimSpace(strings.TrimPrefix(line, "module"))
					return moduleName, dir, nil
				}
			}
			return "", "", fmt.Errorf("module declaration not found in go.mod")
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return "", "", fmt.Errorf("go.mod not found")
}
