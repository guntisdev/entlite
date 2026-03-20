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
)

func Generate(inputFilePath string, parsedEntities []schema.Entity, entityImports map[string]internalParser.ImportInfo, sqlDialect sqlc.SQLDialect) (string, error) {
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

	entityMap := make(map[string]schema.Entity)
	for _, entity := range parsedEntities {
		entityMap[entity.Name] = entity
	}

	createParamsStructs := make(map[string]*ast.StructType)
	createFuncs := make(map[string]*ast.FuncDecl)

	updateParamsStructs := make(map[string]*ast.StructType)
	updateFuncs := make(map[string]*ast.FuncDecl)

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
						if strings.HasPrefix(typeSpec.Name.Name, "Update") && strings.HasSuffix(typeSpec.Name.Name, "Params") {
							updateParamsStructs[typeSpec.Name.Name] = structType
						}
					}
				}
			}
		case *ast.FuncDecl:
			// Find Create{Entity} functions
			if strings.HasPrefix(d.Name.Name, "Create") && d.Recv != nil {
				createFuncs[d.Name.Name] = d
			}
			if strings.HasPrefix(d.Name.Name, "Update") && d.Recv != nil {
				updateFuncs[d.Name.Name] = d
			}
		}
	}

	var sb strings.Builder

	packageName := filepath.Base(filepath.Dir(absInputDir))
	sb.WriteString(fmt.Sprintf("package %s\n\n", packageName))

	needsContextImport := false
	needsSQLImport := false
	needsFmtImport := false

	for structName := range createParamsStructs {
		entityName := strings.TrimSuffix(strings.TrimPrefix(structName, "Create"), "Params")
		if entity, ok := entityMap[entityName]; ok {
			if hasDefaultFuncFields(entity) {
				needsContextImport = true
			}
		}
		if structType, ok := createParamsStructs[structName]; ok {
			if usesSQLTypes(structType) {
				needsSQLImport = true
			}
		}
	}

	if filepath.Base(inputFilePath) == "queries.sql.go" {
		for _, entity := range parsedEntities {
			if hasValidateField(entity) {
				needsFmtImport = true
			}
		}
	}

	sb.WriteString("import (\n")
	// TODO figure out imports
	if needsContextImport {
		sb.WriteString("\t\"context\"\n")
	}
	if needsSQLImport {
		sb.WriteString("\t\"database/sql\"\n")
	}
	if needsFmtImport {
		sb.WriteString("\t\"fmt\"\n")
	}
	if filepath.Base(inputFilePath) == "queries.sql.go" {
		sb.WriteString("\t\"math\"\n")
		// sb.WriteString("\t\"time\"\n")
		sb.WriteString("\t\"google.golang.org/protobuf/types/known/timestamppb\"\n")
	}

	// we need these imports only for overriden queries
	if filepath.Base(inputFilePath) == "queries.sql.go" {
		basePath := filepath.Dir(filepath.Dir(importPath)) // Remove "/db/internal"
		sb.WriteString(fmt.Sprintf("\tpb \"%s/pb\"\n", basePath))

		// Sort keys for consistent output
		keys := make([]string, 0, len(entityImports))
		for key := range entityImports {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			importInfo := entityImports[key]
			sb.WriteString(fmt.Sprintf("\t\"%s\"\n", importInfo.Path))
		}
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
							if entity, ok := entityMap[entityName]; ok {
								// TODO - need to include also for sql type conversion
								if hasDefaultFuncFields(entity) || hasValidateField(entity) {
									// Generate custom struct without DefaultFunc fields, also put Validate
									sb.WriteString(generateCreateStruct(s.Name.Name, createParamsStructs[s.Name.Name], entity))
									continue
								}
							}
						}

						if strings.HasPrefix(s.Name.Name, "Update") && strings.HasSuffix(s.Name.Name, "Params") {
							entityName := strings.TrimSuffix(strings.TrimPrefix(s.Name.Name, "Update"), "Params")
							if entity, ok := entityMap[entityName]; ok && hasDefaultFuncAndNoImmutable(entity) {
								sb.WriteString(generateUpdateStruct(s.Name.Name, updateParamsStructs[s.Name.Name], entity))
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
				// Special handling for New function to return wrapped Queries type
				if d.Name.Name == "New" {
					sb.WriteString(fmt.Sprintf("func %s(db DBTX) *Queries { return (*Queries)(%s.%s(db)) }\n", d.Name.Name, inputPackageName, d.Name.Name))
				} else {
					sb.WriteString(fmt.Sprintf("var %s = %s.%s\n", d.Name.Name, inputPackageName, d.Name.Name))
				}
			} else if d.Recv != nil && strings.HasPrefix(d.Name.Name, "Create") {
				entityName := strings.TrimPrefix(d.Name.Name, "Create")
				if entity, ok := entityMap[entityName]; ok && hasDefaultFuncFields(entity) {
					sb.WriteString(generateCreateMethod(d, entity, inputPackageName, sqlDialect))
					continue
				}
			} else if d.Recv != nil && strings.HasPrefix(d.Name.Name, "Update") {
				entityName := strings.TrimPrefix(d.Name.Name, "Update")
				if entity, ok := entityMap[entityName]; ok && hasDefaultFuncAndNoImmutable(entity) {
					sb.WriteString(generateUpdateMethod(d, entity, inputPackageName, sqlDialect))
					continue
				}
			} else if d.Recv != nil && strings.HasPrefix(d.Name.Name, "Get") {
				entityName := strings.TrimPrefix(d.Name.Name, "Get")
				if entity, ok := entityMap[entityName]; ok {
					sb.WriteString(generateGetMethod(d, entity, inputPackageName, sqlDialect))
					continue
				}
			} else if d.Recv != nil && strings.HasPrefix(d.Name.Name, "List") {
				entityName := strings.TrimPrefix(d.Name.Name, "List")
				if entity, ok := entityMap[entityName]; ok {
					sb.WriteString(generateListMethod(d, entity, inputPackageName, sqlDialect))
					continue
				}
			} else if d.Recv != nil && strings.HasPrefix(d.Name.Name, "Delete") {
				entityName := strings.TrimPrefix(d.Name.Name, "Delete")
				if entity, ok := entityMap[entityName]; ok {
					sb.WriteString(generateDeleteMethod(d, entity, inputPackageName, sqlDialect))
					continue
				}
			}
		}
	}

	if filepath.Base(inputFilePath) == "queries.sql.go" {
		sb.WriteString(generateConverterFunctions())
	}

	return sb.String(), nil
}
