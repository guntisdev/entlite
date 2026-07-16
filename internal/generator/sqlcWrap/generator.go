package sqlcwrap

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"sort"
	"strings"

	internalParser "github.com/guntisdev/entlite/internal/parser"
	"github.com/guntisdev/entlite/internal/schema"
	"github.com/guntisdev/entlite/internal/util"
	"github.com/guntisdev/entlite/pkg/entlite/permissions"
)

// FileType represents the type of sqlc-generated file being wrapped
type FileType int

const (
	FileTypeUnknown FileType = iota
	FileTypeQuery            // *.sql.go files (queries.sql.go, etc.)
	FileTypeModel            // models.go
	FileTypeDB               // db.go
)

func detectFileType(filename string) FileType {
	base := filepath.Base(filename)
	if base == "models.go" {
		return FileTypeModel
	}
	if base == "db.go" {
		return FileTypeDB
	}
	if strings.HasSuffix(base, ".sql.go") {
		return FileTypeQuery
	}
	return FileTypeUnknown
}

func Generate(inputFilePath string, pbDir string, parsedEntities []schema.Entity, entityImports map[string]internalParser.ImportInfo, sqlDialect schema.SQLDialect) (string, error) {
	fileType := detectFileType(inputFilePath)

	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, inputFilePath, nil, parser.ParseComments)
	if err != nil {
		return "", fmt.Errorf("failed to parse file: %w", err)
	}

	inputPackageName := node.Name.Name
	absInputDir, _ := filepath.Abs(filepath.Dir(inputFilePath))
	importPath, err := util.PathToImport(inputFilePath)
	if err != nil {
		return "", fmt.Errorf("failed to convert path to import: %w", err)
	}

	pbImportPath := ""
	if pbDir != "" {
		pbImportPath, err = util.PathToImport(pbDir)
		if err != nil {
			return "", fmt.Errorf("failed to convert pb path to import: %w", err)
		}
	}

	entityMap := make(map[string]schema.Entity)
	for _, entity := range parsedEntities {
		entityMap[entity.Name] = entity
	}

	ctx := &generationContext{
		fileType:            fileType,
		inputFilePath:       inputFilePath,
		inputPackageName:    inputPackageName,
		absInputDir:         absInputDir,
		importPath:          importPath,
		pbImportPath:        pbImportPath,
		node:                node,
		entityMap:           entityMap,
		parsedEntities:      parsedEntities,
		entityImports:       entityImports,
		sqlDialect:          sqlDialect,
		createParamsStructs: make(map[string]*ast.StructType),
		updateParamsStructs: make(map[string]*ast.StructType),
	}

	ctx.collectDeclarations()

	// Generate the body first so imports can be derived from what is actually
	// referenced, rather than guessed up front (which left unused imports for
	// files that emit only passthrough forwarders, e.g. custom.sql.go).
	var body string
	switch fileType {
	case FileTypeQuery:
		body = ctx.generateQueryFileDeclarations()
	case FileTypeModel:
		body = ctx.generateModelFileDeclarations()
	case FileTypeDB:
		body = ctx.generateDBFileDeclarations()
	default:
		body = ctx.generateGenericDeclarations()
	}

	var sb strings.Builder
	packageName := filepath.Base(filepath.Dir(absInputDir))
	sb.WriteString(fmt.Sprintf("package %s\n\n", packageName))
	sb.WriteString(ctx.generateImports(body))
	sb.WriteString(body)

	return sb.String(), nil
}

// generationContext holds all the context needed for generating wrapped code
type generationContext struct {
	fileType            FileType
	inputFilePath       string
	inputPackageName    string
	absInputDir         string
	importPath          string
	pbImportPath        string
	node                *ast.File
	entityMap           map[string]schema.Entity
	parsedEntities      []schema.Entity
	entityImports       map[string]internalParser.ImportInfo
	sqlDialect          schema.SQLDialect
	createParamsStructs map[string]*ast.StructType
	updateParamsStructs map[string]*ast.StructType
}

func (ctx *generationContext) collectDeclarations() {
	for _, decl := range ctx.node.Decls {
		switch d := decl.(type) {
		case *ast.GenDecl:
			for _, spec := range d.Specs {
				if typeSpec, ok := spec.(*ast.TypeSpec); ok {
					if structType, ok := typeSpec.Type.(*ast.StructType); ok {
						if strings.HasPrefix(typeSpec.Name.Name, "Create") && strings.HasSuffix(typeSpec.Name.Name, "Params") {
							ctx.createParamsStructs[typeSpec.Name.Name] = structType
						}
						if strings.HasPrefix(typeSpec.Name.Name, "Update") && strings.HasSuffix(typeSpec.Name.Name, "Params") {
							ctx.updateParamsStructs[typeSpec.Name.Name] = structType
						}
					}
				}
			}
		}
	}
}

// generateImports emits only the imports actually referenced by body. Candidates
// come from the sqlc-generated source file's own imports (so passthrough
// forwarders can name any type the source used), the schema/entity imports
// (validation and default-func helpers), and the proto packages used by model
// converters. The internal package is always imported.
func (ctx *generationContext) generateImports(body string) string {
	type importSpec struct {
		name  string // selector used in code, e.g. "time", "sql", "logic"
		alias string // explicit alias to emit; "" when the default name suffices
		path  string
	}

	seen := make(map[string]bool)
	var specs []importSpec
	add := func(name, alias, path string) {
		if path == "" || seen[path] {
			return
		}
		seen[path] = true
		specs = append(specs, importSpec{name: name, alias: alias, path: path})
	}

	aliasFor := func(name, path string) string {
		if name != filepath.Base(path) {
			return name
		}
		return ""
	}

	// Imports carried over from the sqlc-generated source file.
	for _, imp := range ctx.node.Imports {
		path := strings.Trim(imp.Path.Value, `"`)
		name := filepath.Base(path)
		if imp.Name != nil {
			name = imp.Name.Name
		}
		add(name, aliasFor(name, path), path)
	}

	// Imports needed by DSL-based wrappers (validation funcs, default funcs, etc.).
	keys := make([]string, 0, len(ctx.entityImports))
	for key := range ctx.entityImports {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		info := ctx.entityImports[key]
		add(info.Name, aliasFor(info.Name, info.Path), info.Path)
	}

	// Proto packages used by model converters.
	add("timestamppb", "", "google.golang.org/protobuf/types/known/timestamppb")
	if ctx.pbImportPath != "" {
		add("pb", "pb", ctx.pbImportPath)
	}
	// context/fmt are used by generated wrappers even if the source file omits them.
	add("context", "", "context")
	add("fmt", "", "fmt")

	used := make([]importSpec, 0, len(specs))
	for _, s := range specs {
		if usesPackage(body, s.name) {
			used = append(used, s)
		}
	}
	sort.Slice(used, func(i, j int) bool { return used[i].path < used[j].path })

	var sb strings.Builder
	sb.WriteString("import (\n")
	for _, s := range used {
		if s.alias != "" {
			sb.WriteString(fmt.Sprintf("\t%s \"%s\"\n", s.alias, s.path))
		} else {
			sb.WriteString(fmt.Sprintf("\t\"%s\"\n", s.path))
		}
	}
	// The internal package is always referenced by wrappers/forwarders.
	sb.WriteString(fmt.Sprintf("\t%s \"%s\"\n", ctx.inputPackageName, ctx.importPath))
	sb.WriteString(")\n\n")

	return sb.String()
}

func (ctx *generationContext) generateGenericDeclarations() string {
	var sb strings.Builder

	for _, decl := range ctx.node.Decls {
		switch d := decl.(type) {
		case *ast.GenDecl:
			ctx.reexportGenDecl(&sb, d)
		case *ast.FuncDecl:
			ctx.reexportFunc(&sb, d)
		}
	}

	return sb.String()
}

func (ctx *generationContext) reexportGenDecl(sb *strings.Builder, decl *ast.GenDecl) {
	for _, spec := range decl.Specs {
		switch s := spec.(type) {
		case *ast.TypeSpec:
			if s.Name.IsExported() {
				sb.WriteString(fmt.Sprintf("type %s = %s.%s\n", s.Name.Name, ctx.inputPackageName, s.Name.Name))
			}
		case *ast.ValueSpec:
			for _, name := range s.Names {
				if name.IsExported() {
					if decl.Tok == token.CONST {
						sb.WriteString(fmt.Sprintf("const %s = %s.%s\n", name.Name, ctx.inputPackageName, name.Name))
					} else {
						sb.WriteString(fmt.Sprintf("var %s = %s.%s\n", name.Name, ctx.inputPackageName, name.Name))
					}
				}
			}
		}
	}
}

func (ctx *generationContext) reexportFunc(sb *strings.Builder, funcDecl *ast.FuncDecl) {
	if funcDecl.Name.IsExported() && funcDecl.Recv == nil {
		sb.WriteString(fmt.Sprintf("var %s = %s.%s\n", funcDecl.Name.Name, ctx.inputPackageName, funcDecl.Name.Name))
	}
}

// generateDBFileDeclarations handles db.go with special Queries type handling
func (ctx *generationContext) generateDBFileDeclarations() string {
	var sb strings.Builder

	for _, decl := range ctx.node.Decls {
		switch d := decl.(type) {
		case *ast.GenDecl:
			ctx.processDBGenDecl(&sb, d)
		case *ast.FuncDecl:
			ctx.processDBFunc(&sb, d)
		}
	}

	return sb.String()
}

// processDBGenDecl processes type declarations for db.go
func (ctx *generationContext) processDBGenDecl(sb *strings.Builder, decl *ast.GenDecl) {
	for _, spec := range decl.Specs {
		switch s := spec.(type) {
		case *ast.TypeSpec:
			if !s.Name.IsExported() {
				continue
			}

			// For Queries type, use a proper type (not alias) so we can add methods
			if s.Name.Name == "Queries" {
				sb.WriteString(fmt.Sprintf("type %s %s.%s\n", s.Name.Name, ctx.inputPackageName, s.Name.Name))
			} else {
				sb.WriteString(fmt.Sprintf("type %s = %s.%s\n", s.Name.Name, ctx.inputPackageName, s.Name.Name))
			}

		case *ast.ValueSpec:
			for _, name := range s.Names {
				if name.IsExported() {
					if decl.Tok == token.CONST {
						sb.WriteString(fmt.Sprintf("const %s = %s.%s\n", name.Name, ctx.inputPackageName, name.Name))
					} else {
						sb.WriteString(fmt.Sprintf("var %s = %s.%s\n", name.Name, ctx.inputPackageName, name.Name))
					}
				}
			}
		}
	}
}

// processDBFunc processes function declarations for db.go
func (ctx *generationContext) processDBFunc(sb *strings.Builder, funcDecl *ast.FuncDecl) {
	if funcDecl.Name.IsExported() && funcDecl.Recv == nil {
		// Special handling for New function to return wrapped Queries type
		if funcDecl.Name.Name == "New" {
			sb.WriteString(fmt.Sprintf("func %s(db DBTX) *Queries { return (*Queries)(%s.%s(db)) }\n",
				funcDecl.Name.Name, ctx.inputPackageName, funcDecl.Name.Name))
		} else {
			sb.WriteString(fmt.Sprintf("var %s = %s.%s\n", funcDecl.Name.Name, ctx.inputPackageName, funcDecl.Name.Name))
		}
	}
}

// generateQueryFileDeclarations handles *.sql.go files with query method overrides
func (ctx *generationContext) generateQueryFileDeclarations() string {
	var sb strings.Builder

	for _, decl := range ctx.node.Decls {
		switch d := decl.(type) {
		case *ast.GenDecl:
			ctx.processQueryGenDecl(&sb, d)
		case *ast.FuncDecl:
			ctx.processQueryFunc(&sb, d)
		}
	}

	return sb.String()
}

func (ctx *generationContext) processQueryGenDecl(sb *strings.Builder, decl *ast.GenDecl) {
	for _, spec := range decl.Specs {
		switch s := spec.(type) {
		case *ast.TypeSpec:
			if !s.Name.IsExported() {
				continue
			}

			if strings.HasPrefix(s.Name.Name, "Create") && strings.HasSuffix(s.Name.Name, "Params") {
				entityName := strings.TrimSuffix(strings.TrimPrefix(s.Name.Name, "Create"), "Params")
				if entity, ok := ctx.entityMap[entityName]; ok {
					sb.WriteString(generateCreateStruct(s.Name.Name, ctx.createParamsStructs[s.Name.Name], entity))
					continue
				}
			}

			if strings.HasPrefix(s.Name.Name, "Update") && strings.HasSuffix(s.Name.Name, "Params") {
				entityName := strings.TrimSuffix(strings.TrimPrefix(s.Name.Name, "Update"), "Params")
				if entity, ok := ctx.entityMap[entityName]; ok {
					sb.WriteString(generateUpdateStruct(s.Name.Name, ctx.updateParamsStructs[s.Name.Name], entity))
					continue
				}
			}

			if s.Name.Name == "Queries" {
				sb.WriteString(fmt.Sprintf("type %s %s.%s\n", s.Name.Name, ctx.inputPackageName, s.Name.Name))
			} else {
				sb.WriteString(fmt.Sprintf("type %s = %s.%s\n", s.Name.Name, ctx.inputPackageName, s.Name.Name))
			}

		case *ast.ValueSpec:
			for _, name := range s.Names {
				if name.IsExported() {
					if decl.Tok == token.CONST {
						sb.WriteString(fmt.Sprintf("const %s = %s.%s\n", name.Name, ctx.inputPackageName, name.Name))
					} else {
						sb.WriteString(fmt.Sprintf("var %s = %s.%s\n", name.Name, ctx.inputPackageName, name.Name))
					}
				}
			}
		}
	}
}

func (ctx *generationContext) processQueryFunc(sb *strings.Builder, funcDecl *ast.FuncDecl) {
	if funcDecl.Name.IsExported() && funcDecl.Recv == nil {
		if funcDecl.Name.Name == "New" {
			sb.WriteString(fmt.Sprintf("func %s(db DBTX) *Queries { return (*Queries)(%s.%s(db)) }\n",
				funcDecl.Name.Name, ctx.inputPackageName, funcDecl.Name.Name))
		} else {
			sb.WriteString(fmt.Sprintf("var %s = %s.%s\n", funcDecl.Name.Name, ctx.inputPackageName, funcDecl.Name.Name))
		}
		return
	}

	if funcDecl.Recv != nil {
		// CRUD method overrides
		if strings.HasPrefix(funcDecl.Name.Name, "Create") {
			entityName := strings.TrimPrefix(funcDecl.Name.Name, "Create")
			if entity, ok := ctx.entityMap[entityName]; ok {
				sb.WriteString(generateCreateQuery(funcDecl, entity, ctx.inputPackageName, ctx.sqlDialect))
				return
			}
		}
		if strings.HasPrefix(funcDecl.Name.Name, "Update") {
			entityName := strings.TrimPrefix(funcDecl.Name.Name, "Update")
			if entity, ok := ctx.entityMap[entityName]; ok {
				sb.WriteString(generateUpdateQuery(funcDecl, entity, ctx.inputPackageName, ctx.sqlDialect))
				return
			}
		}
		if strings.HasPrefix(funcDecl.Name.Name, "Get") {
			if entity, ok := ctx.findEntityForGetMethod(funcDecl.Name.Name); ok {
				sb.WriteString(generateGetQuery(funcDecl, entity, ctx.inputPackageName, ctx.sqlDialect))
				return
			}
		}
		if strings.HasPrefix(funcDecl.Name.Name, "List") {
			if entity, ok := ctx.findEntityForListMethod(funcDecl.Name.Name); ok {
				sb.WriteString(generateListQuery(funcDecl, entity, ctx.inputPackageName, ctx.sqlDialect))
				return
			}
		}
		if strings.HasPrefix(funcDecl.Name.Name, "Delete") {
			entityName := strings.TrimPrefix(funcDecl.Name.Name, "Delete")
			if entity, ok := ctx.entityMap[entityName]; ok {
				sb.WriteString(generateDeleteQuery(funcDecl, entity, ctx.inputPackageName, ctx.sqlDialect))
				return
			}
		}

		// Baseline: any exported method that isn't a recognized DSL query (custom
		// hand-written queries, or a DSL query whose entity lookup failed) is
		// re-exposed as a thin passthrough so the wrapped Queries type keeps the
		// full API surface of the underlying sqlc Queries.
		if funcDecl.Name.IsExported() {
			sb.WriteString(ctx.generateForwarder(funcDecl))
		}
	}
}

// generateForwarder emits a thin method that forwards to the underlying sqlc
// Queries verbatim, qualifying any package-local types with the internal package.
func (ctx *generationContext) generateForwarder(funcDecl *ast.FuncDecl) string {
	pkg := ctx.inputPackageName

	var params []string
	var args []string
	unnamed := 0
	if funcDecl.Type.Params != nil {
		for _, p := range funcDecl.Type.Params.List {
			typ := qualifyType(p.Type, pkg)
			if len(p.Names) == 0 {
				name := fmt.Sprintf("a%d", unnamed)
				unnamed++
				params = append(params, name+" "+typ)
				args = append(args, name)
				continue
			}
			for _, n := range p.Names {
				params = append(params, n.Name+" "+typ)
				args = append(args, n.Name)
			}
		}
	}

	var results []string
	if funcDecl.Type.Results != nil {
		for _, r := range funcDecl.Type.Results.List {
			typ := qualifyType(r.Type, pkg)
			count := 1
			if len(r.Names) > 0 {
				count = len(r.Names)
			}
			for i := 0; i < count; i++ {
				results = append(results, typ)
			}
		}
	}

	resultStr := ""
	switch len(results) {
	case 0:
	case 1:
		resultStr = " " + results[0]
	default:
		resultStr = " (" + strings.Join(results, ", ") + ")"
	}

	call := fmt.Sprintf("(*%s.Queries)(q).%s(%s)", pkg, funcDecl.Name.Name, strings.Join(args, ", "))
	// A method returning the underlying *Queries (e.g. WithTx) must hand back the
	// wrapped type, not the internal one, or the wrapper is lost mid-chain.
	if len(results) == 1 && results[0] == "*Queries" {
		call = fmt.Sprintf("(*Queries)(%s)", call)
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("func (q *Queries) %s(%s)%s {\n", funcDecl.Name.Name, strings.Join(params, ", "), resultStr))
	if len(results) == 0 {
		sb.WriteString(fmt.Sprintf("\t%s\n", call))
	} else {
		sb.WriteString(fmt.Sprintf("\treturn %s\n", call))
	}
	sb.WriteString("}\n\n")
	return sb.String()
}

// search for Get[entityname]By[paramnames]
func (ctx *generationContext) findEntityForGetMethod(methodName string) (schema.Entity, bool) {
	entitySuffix := strings.TrimPrefix(methodName, "Get")

	bestMatchName := ""
	for _, entity := range ctx.parsedEntities {
		entityName := entity.Name
		if !strings.HasPrefix(entitySuffix, entityName) {
			continue
		}

		remainder := entitySuffix[len(entityName):]
		if remainder == "" || strings.HasPrefix(remainder, "By") {
			if len(entityName) > len(bestMatchName) {
				bestMatchName = entityName
			}
		}
	}

	if bestMatchName != "" {
		return ctx.entityMap[bestMatchName], true
	}

	return schema.Entity{}, false
}

// search for List[entityname]By[paramnames] and List[entityname]FilterBy[filternames]
func (ctx *generationContext) findEntityForListMethod(methodName string) (schema.Entity, bool) {
	entitySuffix := strings.TrimPrefix(methodName, "List")

	bestMatchName := ""
	for _, entity := range ctx.parsedEntities {
		entityName := entity.Name
		if !strings.HasPrefix(entitySuffix, entityName) {
			continue
		}

		remainder := entitySuffix[len(entityName):]
		if remainder == "" || strings.HasPrefix(remainder, "By") || strings.HasPrefix(remainder, "FilterBy") {
			if len(entityName) > len(bestMatchName) {
				bestMatchName = entityName
			}
		}
	}

	if bestMatchName != "" {
		return ctx.entityMap[bestMatchName], true
	}

	return schema.Entity{}, false
}

func (ctx *generationContext) generateModelFileDeclarations() string {
	var sb strings.Builder

	for _, decl := range ctx.node.Decls {
		if genDecl, ok := decl.(*ast.GenDecl); ok {
			for _, spec := range genDecl.Specs {
				if typeSpec, ok := spec.(*ast.TypeSpec); ok {
					if !typeSpec.Name.IsExported() {
						continue
					}

					isEntityModel := false
					for _, entity := range ctx.parsedEntities {
						if typeSpec.Name.Name == entity.Name {
							isEntityModel = true
							break
						}
					}

					if !isEntityModel {
						sb.WriteString(fmt.Sprintf("type %s = %s.%s\n", typeSpec.Name.Name, ctx.inputPackageName, typeSpec.Name.Name))
					}
				}
			}
		}
	}

	sb.WriteString("\n")

	for _, entity := range ctx.parsedEntities {
		sb.WriteString(ctx.generateEntityModel(entity))
		sb.WriteString("\n")
	}

	for _, entity := range ctx.parsedEntities {
		sb.WriteString(ctx.generateModelConverters(entity))
		sb.WriteString("\n")
	}

	for _, entity := range ctx.parsedEntities {
		sb.WriteString(ctx.generateProtoConverter(entity))
		sb.WriteString("\n")
	}

	return sb.String()
}

func (ctx *generationContext) generateEntityModel(entity schema.Entity) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("type %s struct {\n", entity.Name))

	for _, field := range entity.Fields {
		fieldName := toDBFieldName(field)
		goType := fieldToGoType(field)
		sb.WriteString(fmt.Sprintf("\t%s %s `json:\"%s\"`\n", fieldName, goType, field.Name))
	}

	sb.WriteString("}\n")
	return sb.String()
}

func (ctx *generationContext) generateModelConverters(entity schema.Entity) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("func (m *%s) %sToSQL() *%s.%s {\n", entity.Name, entity.Name, ctx.inputPackageName, entity.Name))
	sb.WriteString("\tif m == nil {\n\t\treturn nil\n\t}\n\n")
	sb.WriteString(fmt.Sprintf("\treturn &%s.%s{\n", ctx.inputPackageName, entity.Name))

	for _, field := range entity.Fields {
		fieldName := toDBFieldName(field)
		convertedValue := sqlToGo(field, "m."+fieldName, ctx.sqlDialect)
		sb.WriteString(fmt.Sprintf("\t\t%s: %s,\n", fieldName, convertedValue))
	}

	sb.WriteString("\t}\n}\n\n")

	sb.WriteString(fmt.Sprintf("func %sFromSQL(db *%s.%s) *%s {\n", entity.Name, ctx.inputPackageName, entity.Name, entity.Name))
	sb.WriteString("\tif db == nil {\n\t\treturn nil\n\t}\n\n")
	sb.WriteString(fmt.Sprintf("\treturn &%s{\n", entity.Name))

	for _, field := range entity.Fields {
		fieldName := toDBFieldName(field)
		convertedValue := goFromSQL(field, "db."+fieldName, ctx.sqlDialect)
		sb.WriteString(fmt.Sprintf("\t\t%s: %s,\n", fieldName, convertedValue))
	}

	sb.WriteString("\t}\n}\n")

	return sb.String()
}
func (ctx *generationContext) generateProtoConverter(entity schema.Entity) string {
	var sb strings.Builder
	protoPackage := "pb"

	sb.WriteString(fmt.Sprintf("// ToProto converts %s to proto format\n", entity.Name))
	sb.WriteString(fmt.Sprintf("func (m *%s) ToProto() *%s.%s {\n", entity.Name, protoPackage, entity.Name))
	sb.WriteString("\tif m == nil {\n\t\treturn nil\n\t}\n\n")
	sb.WriteString(fmt.Sprintf("\treturn &%s.%s{\n", protoPackage, entity.Name))

	for _, field := range entity.Fields {
		canRead := (field.Permissions & permissions.ApiRead) != 0
		if !canRead {
			continue
		}

		protoName := toProtoFieldName(field)
		modelName := toDBFieldName(field)

		if field.Type == schema.FieldTypeTime {
			if field.Optional {
				sb.WriteString(fmt.Sprintf("\t\t%s: func() *timestamppb.Timestamp { if m.%s != nil { return timestamppb.New(*m.%s) }; return nil }(),\n", protoName, modelName, modelName))
			} else {
				sb.WriteString(fmt.Sprintf("\t\t%s: timestamppb.New(m.%s),\n", protoName, modelName))
			}
		} else if field.Type == schema.FieldTypeByte && field.Optional {
			// proto uses []byte for optional bytes; unwrap the wrapper's *[]byte.
			sb.WriteString(fmt.Sprintf("\t\t%s: PtrToNullBytes(m.%s),\n", protoName, modelName))
		} else {
			sb.WriteString(fmt.Sprintf("\t\t%s: m.%s,\n", protoName, modelName))
		}
	}

	sb.WriteString("\t}\n}\n")

	return sb.String()
}
