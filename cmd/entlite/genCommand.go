package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/guntisdev/entlite/internal/generator/proto"
	"github.com/guntisdev/entlite/internal/generator/sqlc"
	"github.com/guntisdev/entlite/internal/schema"
	"github.com/guntisdev/entlite/internal/util"
)

// genCommand writes the sqlc and proto contracts from a schema directory.
// usage: entlite gen <schema dir>
func genCommand(args []string) {
	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, "Error: entity directory is required")
		os.Exit(1)
	}

	entityDir := args[0]
	parsedEntities, err := loadEntities(entityDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading entities: %v\n", err)
		os.Exit(1)
	}

	dir, _ := filepath.Abs(entityDir)
	protoDir := filepath.Join(filepath.Dir(dir), "contract", "proto")
	sqlcDir := filepath.Join(filepath.Dir(dir), "contract", "sqlc")
	genDir := filepath.Join(filepath.Dir(dir), "gen")

	dirs := []string{protoDir, sqlcDir, genDir}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating directory %s %v\n", dir, err)
			os.Exit(1)
		}
	}

	// PROTO
	protoEntities := schema.FilterPROTO(parsedEntities)
	if len(protoEntities) > 0 {
		goPackage, err := pbImportPath(filepath.Dir(dir))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed resolving go_package: %v\n", err)
			os.Exit(1)
		}

		if err := proto.Generate(protoEntities, protoDir, goPackage); err != nil {
			fmt.Fprintf(os.Stderr, "Failed generating proto: %v\n", err)
			os.Exit(1)
		}
	}

	// SQLC
	sqlcEntities := schema.FilterSQLC(parsedEntities)
	if len(sqlcEntities) > 0 {
		sqlcYamlPath := filepath.Join(filepath.Dir(dir), "sqlc.yaml")
		sqlcConfig, err := util.GetSqlcConfigFromYaml(sqlcYamlPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed reading sqlc.yaml: %v\n", err)
			os.Exit(1)
		}

		sqlcGenerator := sqlc.NewGenerator(sqlcConfig.Dialect)
		if err := sqlcGenerator.Generate(sqlcEntities, sqlcDir); err != nil {
			fmt.Fprintf(os.Stderr, "Failed generating sqlc: %v\n", err)
			os.Exit(1)
		}
	}
}

func pbImportPath(entDir string) (string, error) {
	pbDir := filepath.Join(entDir, "gen", "pb")

	// buf.gen.yaml is the source of truth
	if bufConfig, err := util.GetBufConfigFromYaml(filepath.Join(entDir, "buf.gen.yaml")); err == nil {
		pbDir = filepath.Join(entDir, bufConfig.ProtoTypesDir)
	}

	return util.PathToImport(pbDir)
}
