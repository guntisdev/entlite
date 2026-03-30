package sqlcwrap

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"sort"
	"strings"

	"github.com/guntisdev/entlite/internal/generator/sqlc"
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

func Generate(inputFilePath string, pbDir string, parsedEntities []schema.Entity, entityImports map[string]internalParser.ImportInfo, sqlDialect sqlc.SQLDialect) (string, error) {
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

	var sb strings.Builder
	packageName := filepath.Base(filepath.Dir(absInputDir))
	sb.WriteString(fmt.Sprintf("package %s\n\n", packageName))
	sb.WriteString(ctx.generateImports())

	switch fileType {
	case FileTypeQuery:
		sb.WriteString(ctx.generateQueryFileDeclarations())
	case FileTypeModel:
		sb.WriteString(ctx.generateModelFileDeclarations())
	case FileTypeDB:
		sb.WriteString(ctx.generateDBFileDeclarations())
	default:
		sb.WriteString(ctx.generateGenericDeclarations())
	}

	if fileType == FileTypeQuery {
		hasTimeField := false
		for _, entity := range ctx.parsedEntities {
			for _, field := range entity.Fields {
				if field.Type == schema.FieldTypeTime {
					hasTimeField = true
					break
				}
			}
			if hasTimeField {
				break
			}
		}
		sb.WriteString(generateConverterFunctions(hasTimeField))
	}

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
	sqlDialect          sqlc.SQLDialect
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

func (ctx *generationContext) generateImports() string {
	var sb strings.Builder
	sb.WriteString("import (\n")

	needsContext := false
	needsSQL := false
	needsFmt := false

	// Query files always need context import (all methods have ctx context.Context parameter)
	if ctx.fileType == FileTypeQuery {
		needsContext = true
	}

	for structName := range ctx.createParamsStructs {
		entityName := strings.TrimSuffix(strings.TrimPrefix(structName, "Create"), "Params")
		if entity, ok := ctx.entityMap[entityName]; ok {
			if hasDefaultFuncFields(entity) {
				needsContext = true
			}
		}
		if structType, ok := ctx.createParamsStructs[structName]; ok {
			if usesSQLTypes(structType) {
				needsSQL = true
			}
		}
	}

	if ctx.fileType == FileTypeQuery {
		for _, entity := range ctx.parsedEntities {
			if hasValidateField(entity) {
				needsFmt = true
			}
		}
	}

	if needsContext {
		sb.WriteString("\t\"context\"\n")
	}
	if needsSQL || ctx.fileType == FileTypeQuery {
		sb.WriteString("\t\"database/sql\"\n")
	}
	if needsFmt {
		sb.WriteString("\t\"fmt\"\n")
	}

	switch ctx.fileType {
	case FileTypeQuery:
		sb.WriteString("\t\"math\"\n")

		hasTimeField := false
		for _, entity := range ctx.parsedEntities {
			for _, field := range entity.Fields {
				if field.Type == schema.FieldTypeTime {
					hasTimeField = true
					break
				}
			}
			if hasTimeField {
				break
			}
		}
		if hasTimeField {
			sb.WriteString("\t\"google.golang.org/protobuf/types/known/timestamppb\"\n")
		}

		keys := make([]string, 0, len(ctx.entityImports))
		for key := range ctx.entityImports {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			importInfo := ctx.entityImports[key]
			sb.WriteString(fmt.Sprintf("\t\"%s\"\n", importInfo.Path))
		}

	case FileTypeModel:
		hasTimeField := false
		for _, entity := range ctx.parsedEntities {
			for _, field := range entity.Fields {
				if field.Type == schema.FieldTypeTime {
					hasTimeField = true
					break
				}
			}
			if hasTimeField {
				break
			}
		}
		if hasTimeField {
			sb.WriteString("\t\"time\"\n")
			sb.WriteString("\t\"google.golang.org/protobuf/types/known/timestamppb\"\n")
		}
		if ctx.pbImportPath != "" {
			sb.WriteString(fmt.Sprintf("\tpb \"%s\"\n", ctx.pbImportPath))
		}
	}

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
				sb.WriteString(generateCreateMethod(funcDecl, entity, ctx.inputPackageName, ctx.sqlDialect))
				return
			}
		}
		if strings.HasPrefix(funcDecl.Name.Name, "Update") {
			entityName := strings.TrimPrefix(funcDecl.Name.Name, "Update")
			if entity, ok := ctx.entityMap[entityName]; ok {
				sb.WriteString(generateUpdateMethod(funcDecl, entity, ctx.inputPackageName, ctx.sqlDialect))
				return
			}
		}
		if strings.HasPrefix(funcDecl.Name.Name, "Get") {
			entityName := strings.TrimPrefix(funcDecl.Name.Name, "Get")
			if entity, ok := ctx.entityMap[entityName]; ok {
				sb.WriteString(generateGetMethod(funcDecl, entity, ctx.inputPackageName, ctx.sqlDialect))
				return
			}
		}
		if strings.HasPrefix(funcDecl.Name.Name, "List") {
			entityName := strings.TrimPrefix(funcDecl.Name.Name, "List")
			if entity, ok := ctx.entityMap[entityName]; ok {
				sb.WriteString(generateListMethod(funcDecl, entity, ctx.inputPackageName, ctx.sqlDialect))
				return
			}
		}
		if strings.HasPrefix(funcDecl.Name.Name, "Delete") {
			entityName := strings.TrimPrefix(funcDecl.Name.Name, "Delete")
			if entity, ok := ctx.entityMap[entityName]; ok {
				sb.WriteString(generateDeleteMethod(funcDecl, entity, ctx.inputPackageName, ctx.sqlDialect))
				return
			}
		}
	}
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

		fieldName := toDBFieldName(field)

		if field.Type == schema.FieldTypeTime {
			if field.Optional {
				sb.WriteString(fmt.Sprintf("\t\t%s: func() *timestamppb.Timestamp { if m.%s != nil { return timestamppb.New(*m.%s) }; return nil }(),\n", fieldName, fieldName, fieldName))
			} else {
				sb.WriteString(fmt.Sprintf("\t\t%s: timestamppb.New(m.%s),\n", fieldName, fieldName))
			}
		} else {
			sb.WriteString(fmt.Sprintf("\t\t%s: m.%s,\n", fieldName, fieldName))
		}
	}

	sb.WriteString("\t}\n}\n")

	return sb.String()
}
