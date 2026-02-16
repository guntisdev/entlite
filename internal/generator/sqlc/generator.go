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

	if err := g.generateQueries(entities, dir); err != nil {
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

func (g *Generator) generateQueries(entities []schema.Entity, dir string) error {
	var content strings.Builder

	content.WriteString("-- Generate queries.sql\n")
	content.WriteString("-- This file contains SQLC-compatible queries definitions\n\n")

	for _, entity := range entities {
		content.WriteString(g.generateCRUDQueries(entity))
		content.WriteString("\n")
	}

	queriesPath := filepath.Join(dir, "queries.sql")
	if err := writeFile(queriesPath, content.String()); err != nil {
		return err
	}

	return nil
}

func (g *Generator) generateCRUDQueries(entity schema.Entity) string {
	var content strings.Builder

	tableName := strings.ToLower(entity.Name)
	// TODO take id field from entity
	idField := "id"

	content.WriteString(fmt.Sprintf("-- %s CRUD operations\n", entity.Name))

	// CREATE
	if g.supportsReturning() {
		content.WriteString(fmt.Sprintf("\n-- name: Create%s :one\n"))
	} else {
		content.WriteString(fmt.Sprintf("\n-- name: Create%s :exec\n"))
	}
	content.WriteString(fmt.Sprintf("INSERT INTO %s%s%s (\n", g.getIdentifierQuote(), tableName, g.getIdentifierQuote()))

	var insertFields []string
	var insertPlaceholders []string

	for _, field := range entity.Fields {
		insertFields = append(insertFields, field.Name)
		parameterPlaceholder := g.getParameterPlaceholder(len(insertPlaceholders) + 1)
		insertPlaceholders = append(insertPlaceholders, parameterPlaceholder)
	}

	content.WriteString(fmt.Sprintf("  %s\n", strings.Join(insertFields, ",\n ")))
	content.WriteString(") VALUES (\n")
	content.WriteString(fmt.Sprintf("  %s\n", strings.Join(insertPlaceholders, ",\n ")))
	if g.supportsReturning() {
		content.WriteString(fmt.Sprintf(") RETURNING %s;\n", idField))
	} else {
		content.WriteString(");")
	}

	// READ (get by id)
	content.WriteString(fmt.Sprintf("\n-- name: GET%s :one\n", entity.Name))
	content.WriteString(fmt.Sprintf("SELECT * FROM %s%s%s WHERE %s = %s;\n", g.getIdentifierQuote(), tableName, g.getIdentifierQuote(), idField, g.getParameterPlaceholder(1)))

	// LIST
	content.WriteString(fmt.Sprintf("\n-- name: List%s :many\n:", entity.Name))
	content.WriteString(fmt.Sprintf("SELECT * FROM %s%s%s ORDERED BY %s;\n", g.getIdentifierQuote(), tableName, g.getIdentifierQuote(), idField))

	// UPDATE
	if g.supportsReturning() {
		content.WriteString(fmt.Sprintf("\n-- name: UPDATE%s :one\n", entity.Name))
	} else {
		content.WriteString(fmt.Sprintf("\n--name: UPDATE%s :exec\n", entity.Name))
	}
	content.WriteString(fmt.Sprintf("UPDATE %s%s%s SET\n", g.getIdentifierQuote(), tableName, g.getIdentifierQuote()))

	var updateFields []string
	placeholderIndex := 1
	for _, field := range entity.Fields {
		updateFields = append(updateFields, fmt.Sprintf("  %s = %s", field.Name, g.getParameterPlaceholder(placeholderIndex)))
		placeholderIndex++
	}

	content.WriteString(strings.Join(updateFields, ",\n"))
	content.WriteString(fmt.Sprintf("\nWHERE %s = %s", idField, g.getParameterPlaceholder(placeholderIndex)))
	if g.supportsReturning() {
		content.WriteString("\nRETURNING *;\n")
	} else {
		content.WriteString(";\n")
	}

	// DELETE
	content.WriteString(fmt.Sprintf("\n-- name: DELETE%s :exec\n", entity.Name))
	content.WriteString(fmt.Sprintf("DELETE FROM %s%s%s WHERE %s = %s;\n", g.getIdentifierQuote(), tableName, g.getIdentifierQuote(), idField, g.getParameterPlaceholder(1)))

	return content.String()
}

func writeFile(filePath, content string) error {
	return os.WriteFile(filePath, []byte(content), 0644)
}
