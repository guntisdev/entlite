package main

import (
	"fmt"
	"os"
	"path"

	protovalidate "github.com/guntisdev/entlite/internal/generator/protoValidate"
	"github.com/guntisdev/entlite/internal/util"
)

func protoValidate() {
	entityDir := "./schema"
	parsedEntities, err := loadEntities(entityDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading entities: %v\n", err)
		os.Exit(1)
	}

	bufConfig, err := util.GetBufConfigFromYaml("./buf.gen.yaml")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed reading buf.gen.yaml: %v\n", err)
		os.Exit(1)
	}

	outputDir := bufConfig.ProtoTypesDir

	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Output directory does not exist, must be created by buf %v\n", err)
		os.Exit(1)
	}

	validateImports, err := getValidateImports(entityDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading validate imports: %v\n", err)
		os.Exit(1)
	}

	content, err := protovalidate.Generate(parsedEntities, validateImports)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating proto validation: %v\n", err)
		os.Exit(1)
	}

	outputPath := path.Join(outputDir, "proto_validate.go")
	err = os.WriteFile(outputPath, []byte(content), 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing output file %s: %v\n", outputPath, err)
		os.Exit(1)
	}
}
