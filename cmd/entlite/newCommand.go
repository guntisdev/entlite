package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func newCommand(entityNames []string) {
	if len(entityNames) == 0 {
		fmt.Fprintln(os.Stderr, "Please specify at least one entity name")
		os.Exit(1)
	}

	entDir := "ent"
	schemaDir := filepath.Join(entDir, "schema")
	contractDir := filepath.Join(entDir, "contract")
	genDir := filepath.Join(entDir, "gen")

	for _, dir := range []string{entDir, schemaDir, contractDir, genDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating %s directory: %v\n", dir, err)
			os.Exit(1)
		}
	}

	for _, entityName := range entityNames {
		if err := createEntityFile(entityName, schemaDir); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating %s entity: %v\n", entityName, err)
			os.Exit(1)
		}
	}

	if err := createGenFile(schemaDir); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating generate.go: %v\n", err)
		os.Exit(1)
	}

	if err := createSqlcYamlFile(entDir); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating sqlc.yaml: %v\n", err)
		os.Exit(1)
	}

	if err := createBufYamlFile(entDir); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating buf.yaml: %v\n", err)
		os.Exit(1)
	}

	if err := createBufGenYamlFile(entDir); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating buf.gen.yaml: %v\n", err)
		os.Exit(1)
	}
}

func createEntityFile(entityName string, dir string) error {
	fileName := strings.ToLower(entityName) + ".go"
	filePath := filepath.Join(dir, fileName)

	content := fmt.Sprintf(`package ent

import "github.com/guntisdev/entlite/pkg/entlite"

// %s entity definition
type %s struct {
	entlite.Schema
}

func (%s) Annotations() []entlite.Annotation {
	return []entlite.Annotation{
		entlite.Message(),
		entlite.Service(),
	}
}

func (%s) Fields() []entlite.Field {
	return []entlite.Field{
		entlite.String("name").ProtoField(2),
		// Add more fields here
	}
}
`, entityName, entityName, entityName, entityName)

	return createIfNotExist(filePath, content)
}

func createGenFile(dir string) error {
	content := `//go:generate go run github.com/guntisdev/entlite/cmd/entlite gen .

package ent
`
	filePath := filepath.Join(dir, "generate.go")

	return createIfNotExist(filePath, content)
}

func createSqlcYamlFile(dir string) error {
	content := `version: "2"
sql:
  - schema: "contract/sqlc/schema.sql"
    queries: "contract/sqlc/queries.sql"    
    engine: "postgresql"       # postgresql or sqlite or mysql
    gen:
      go:
        package: "db"
        out: "gen/db"
        emit_json_tags: true  
`

	path := filepath.Join(dir, "sqlc.yaml")
	return createIfNotExist(path, content)
}

func createBufYamlFile(dir string) error {
	content := `version: v2
modules:
  - path: contract/proto
lint:
  use:
    - STANDARD
breaking:
  use:
    - FILE
`

	path := filepath.Join(dir, "buf.yaml")
	return createIfNotExist(path, content)
}

func createBufGenYamlFile(dir string) error {
	content := `version: v2
plugins:
  - remote: buf.build/protocolbuffers/go
    out: gen/pb
    opt: paths=source_relative
  - remote: buf.build/connectrpc/go
    out: gen/pb
    opt: 
      - paths=source_relative
      - package_suffix=
  - remote: buf.build/bufbuild/es
    out: gen/ts
    opt: target=ts
  - remote: buf.build/connectrpc/es
    out: gen/ts
    opt: target=ts
`

	path := filepath.Join(dir, "buf.gen.yaml")
	return createIfNotExist(path, content)
}

func createIfNotExist(path string, content string) error {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)

	if errors.Is(err, os.ErrExist) {
		fmt.Printf("Skipping: %s already exists\n", path)
		return nil
	} else if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer f.Close()

	_, err = f.WriteString(content)
	return err
}
