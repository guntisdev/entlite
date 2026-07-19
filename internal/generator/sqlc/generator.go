package sqlc

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/guntisdev/entlite/internal/schema"
	"github.com/guntisdev/entlite/internal/util"
	"github.com/guntisdev/entlite/pkg/entlite/permissions"
)

type Generator struct {
	sqlDialect schema.SQLDialect
}

func NewGenerator(sqlDialect schema.SQLDialect) *Generator {
	return &Generator{
		sqlDialect: sqlDialect,
	}
}

func (g *Generator) Generate(entities []schema.Entity, dir string) error {
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
	content.WriteString(fmt.Sprintf("CREATE TABLE %s(\n", g.quote(tableName)))

	idField := entity.GetIdField()

	for _, field := range entity.Fields {
		if field.IsID() {
			content.WriteString(g.getIdFieldSQL(idField))
			continue
		}

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
		entityQueries := g.generateCRUDQueries(entity)
		if entityQueries == "" {
			continue
		}
		content.WriteString(entityQueries)
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
	idField := entity.GetIdField()

	var hasCreate bool
	var hasUpdate bool
	var hasDelete bool
	var hasDeleteAll bool
	var getQueries []schema.Query
	var listQueries []schema.Query

	for _, query := range entity.Queries {
		switch query.Type {
		case schema.QueryCreate:
			hasCreate = true
		case schema.QueryUpdate:
			hasUpdate = true
		case schema.QueryDelete:
			hasDelete = true
		case schema.QueryDeleteAll:
			hasDeleteAll = true
		case schema.QueryGetBy:
			getQueries = append(getQueries, query)
		case schema.QueryListBy, schema.QueryListAll:
			listQueries = append(listQueries, query)
		}
	}

	content.WriteString(fmt.Sprintf("-- %s CRUD operations\n", entity.Name))

	// CREATE
	if hasCreate {
		if g.supportsReturning() {
			content.WriteString(fmt.Sprintf("\n-- name: Create%s :one\n", entity.Name))
		} else {
			content.WriteString(fmt.Sprintf("\n-- name: Create%s :execlastid\n", entity.Name))
		}
		content.WriteString(fmt.Sprintf("INSERT INTO %s (\n", g.quote(tableName)))

		var insertFields []string
		var insertPlaceholders []string

		for _, field := range entity.Fields {
			canWrite := (field.Permissions & permissions.DbWrite) != 0
			if !canWrite {
				continue
			}
			if field.IsID() && field.DefaultFunc == nil {
				continue
			}
			insertFields = append(insertFields, " "+field.Name)
			parameterPlaceholder := g.getParameterPlaceholder(len(insertPlaceholders) + 1)
			insertPlaceholders = append(insertPlaceholders, " "+parameterPlaceholder)
		}

		content.WriteString(fmt.Sprintf(" %s\n", strings.Join(insertFields, ",\n ")))
		content.WriteString(") VALUES (\n")
		content.WriteString(fmt.Sprintf(" %s\n", strings.Join(insertPlaceholders, ",\n ")))
		if g.supportsReturning() {
			content.WriteString(fmt.Sprintf(") RETURNING %s;\n", idField.Name))
		} else {
			content.WriteString(");")
		}
	}

	// READ (get by)
	for _, query := range getQueries {
		fieldsStr := util.FieldsToStr(query.Fields)
		queryName := fmt.Sprintf("Get%sBy%s", entity.Name, fieldsStr)
		content.WriteString(fmt.Sprintf("\n-- name: %s :one\n", queryName))
		var whereParts []string
		for i, fieldName := range query.Fields {
			whereParts = append(whereParts, fmt.Sprintf("%s = %s", fieldName, g.getParameterPlaceholder(i+1)))
		}
		// TODO implement !permissions.DbRead
		content.WriteString(fmt.Sprintf("SELECT * FROM %s WHERE %s;\n", g.quote(tableName), strings.Join(whereParts, " AND ")))
	}

	// LIST
	for _, query := range listQueries {
		methodName := util.GenListMethodName(query, entity.Name)
		content.WriteString(fmt.Sprintf("\n-- name: %s :many\n", methodName))
		var whereParts []string
		for _, fieldName := range query.Fields {
			whereParts = append(whereParts, fmt.Sprintf("%s = %s", fieldName, g.namedArg(fieldName)))
		}
		for _, filter := range query.Filters {
			switch filter.Type {
			case schema.QueryFilterEq:
				whereParts = append(whereParts, fmt.Sprintf("%s = %s", filter.Field, g.namedArg(filter.Field)))

			case schema.QueryFilterSearch:
				whereParts = append(whereParts, fmt.Sprintf("%s LIKE %s", filter.Field, g.namedArg(filter.Field)))

			case schema.QueryFilterRange:
				minArg := g.namedArg("min_" + filter.Field)
				maxArg := g.namedArg("max_" + filter.Field)
				whereParts = append(whereParts, fmt.Sprintf("%s BETWEEN %s AND %s", filter.Field, minArg, maxArg))
			}
		}
		// TODO implement !permissions.DbRead
		if len(whereParts) == 0 {
			// ListAll: no filters, no WHERE clause.
			content.WriteString(fmt.Sprintf("SELECT * FROM %s;\n", g.quote(tableName)))
		} else {
			content.WriteString(fmt.Sprintf("SELECT * FROM %s WHERE %s;\n", g.quote(tableName), strings.Join(whereParts, " AND ")))
		}
	}

	// UPDATE
	if hasUpdate {
		if g.supportsReturning() {
			content.WriteString(fmt.Sprintf("\n-- name: Update%s :one\n", entity.Name))
		} else {
			content.WriteString(fmt.Sprintf("\n-- name: Update%s :exec\n", entity.Name))
		}
		content.WriteString(fmt.Sprintf("UPDATE %s SET\n", g.quote(tableName)))

		var updateFields []string
		for _, field := range entity.Fields {
			canWrite := (field.Permissions & permissions.DbWrite) != 0
			if !canWrite {
				continue
			}
			if field.IsID() || field.Immutable {
				continue
			}

			// For non-readable fields (like passwords), use COALESCE with nullable parameter
			canApiRead := (field.Permissions & permissions.ApiRead) != 0
			canApiWrite := (field.Permissions & permissions.ApiWrite) != 0
			acceptOptional := false
			if canApiWrite && (field.DefaultFunc != nil || field.DefaultValue != nil) {
				acceptOptional = true
			}
			var fieldUpdate string
			if !canApiRead || acceptOptional {
				// This makes the field optional in updates - if NULL is passed, keep existing value
				fieldUpdate = fmt.Sprintf("  %s = COALESCE(sqlc.narg('%s'), %s)", field.Name, field.Name, field.Name)
			} else {
				fieldUpdate = fmt.Sprintf("  %s = %s", field.Name, g.namedArg(field.Name))
			}
			updateFields = append(updateFields, fieldUpdate)
		}

		content.WriteString(strings.Join(updateFields, ",\n"))
		content.WriteString(fmt.Sprintf("\nWHERE %s = %s", idField.Name, g.namedArg(idField.Name)))
		if g.supportsReturning() {
			content.WriteString("\nRETURNING *;\n")
		} else {
			content.WriteString(";\n")
		}
	}

	// DELETE
	if hasDelete {
		content.WriteString(fmt.Sprintf("\n-- name: Delete%s :exec\n", entity.Name))
		content.WriteString(fmt.Sprintf("DELETE FROM %s WHERE %s = %s;\n", g.quote(tableName), idField.Name, g.getParameterPlaceholder(1)))
	}

	// DELETE ALL
	if hasDeleteAll {
		content.WriteString(fmt.Sprintf("\n-- name: DeleteAll%s :exec\n", entity.Name))
		content.WriteString(fmt.Sprintf("DELETE FROM %s;\n", g.quote(tableName)))
	}

	return content.String()
}

func writeFile(filePath, content string) error {
	return os.WriteFile(filePath, []byte(content), 0644)
}
