package sqlc

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/guntisdev/entlite/internal/schema"
)

type SQLDialect string

const (
	MySQL      SQLDialect = "mysql"
	SQLite     SQLDialect = "sqlite"
	PostgreSQL SQLDialect = "postgresql"
)

type Generator struct {
	sqlDialect SQLDialect
}

func NewGenerator(sqlDialect SQLDialect) *Generator {
	return &Generator{
		sqlDialect: sqlDialect,
	}
}

func (g *Generator) Generate(entities []schema.Entity, dir string, sqlDialect SQLDialect) error {
	if err := g.generateSchema(entities, dir); err != nil {
		return fmt.Errorf("Failed to generate schema.sql: %w", err)
	}

	if err := g.generateQueries(entities); err != nil {
		return fmt.Errorf("failed to generate queries.sql: %w", err)
	}

	return nil
}

func (g *Generator) generateSchema(entities []schema.Entity, dir string) error {
	var content strings.Builder

	content.WriteString("-- Generated schema.sql\n")
	content.WriteString("-- This file contains table definitions for all entities\n\n")

	for _, entity := range entities {
		content.WriteString(g.generateTableSQL(entity))
		content.WriteString("\n")
	}

	schemaPath := filepath.Join(dir, "schema.sql")
	if err := writeFile(schemaPath, content.String()); err != nil {
		return err
	}

	return nil
}

func (g *Generator) generateTableSQL(entity schema.Entity) string {
	var content strings.Builder

	tableName := strings.ToLower(entity.Name)
	content.WriteString(fmt.Sprintf("-- %s table\n", tableName))
	content.WriteString(fmt.Sprintf("CREATE TABLE %s%s%s(\n", g.getIdentifierQuote(), tableName, g.getIdentifierQuote()))

	content.WriteString(g.getIdFieldSQL())

	for _, field := range entity.Fields {
		content.WriteString(",\n")
		sqlType := g.getSQLType(field.Type)

		content.WriteString(fmt.Sprintf("  %s %s", field.Name, sqlType))

		if field.Unique {
			content.WriteString(" UNIQUE")
		}

		if field.DefaultValue != nil {
			defaultVal := g.formatDefaultValue(field.DefaultValue, field.Type)
			content.WriteString(fmt.Sprintf(" DEFAULT %s", defaultVal))
		}

		if !field.Optional {
			content.WriteString(" NOT NULL")
		}

		// TODO write logic for DefaultFunc, Comment etc
	}

	content.WriteString("\n);\n")
	return content.String()
}

func (g *Generator) generateQueries(entities []schema.Entity) error {

	return nil
}

func writeFile(filePath, content string) error {
	return os.WriteFile(filePath, []byte(content), 0644)
}
