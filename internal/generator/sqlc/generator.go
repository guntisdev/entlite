package sqlc

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/guntisdev/entlite/internal/schema"
	"github.com/guntisdev/entlite/internal/util"
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

	content.WriteString(util.GeneratedSQL)
	content.WriteString("-- Table definitions for all entities\n\n")

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

func writeTableComment(content *strings.Builder, entity schema.Entity, tableName string) {
	if entity.Comment == "" {
		fmt.Fprintf(content, "-- %s table\n", tableName)
		return
	}
	for line := range strings.SplitSeq(entity.Comment, "\n") {
		fmt.Fprintf(content, "-- %s\n", strings.TrimRight(line, "\r"))
	}
}

// writes the sqlc name directive followed by the query comment,
func writeQueryHeader(content *strings.Builder, query schema.Query, queryName, kind string) {
	fmt.Fprintf(content, "\n-- name: %s :%s\n", queryName, kind)
	for line := range strings.SplitSeq(query.Comment, "\n") {
		if line == "" {
			continue
		}
		fmt.Fprintf(content, "-- %s\n", strings.TrimRight(line, "\r"))
	}
}

func writeColumnComment(content *strings.Builder, comment string) {
	if comment == "" {
		return
	}
	for line := range strings.SplitSeq(comment, "\n") {
		fmt.Fprintf(content, "  -- %s\n", strings.TrimRight(line, "\r"))
	}
}

func (g *Generator) generateTableSQL(entity schema.Entity) string {
	var content strings.Builder

	tableName := strings.ToLower(entity.Name)
	writeTableComment(&content, entity, tableName)
	content.WriteString(fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s(\n", g.quote(tableName)))

	idField := entity.GetIdField()

	for _, field := range entity.Fields {
		if field.IsVirtual() {
			continue
		}

		if field.IsID() {
			writeColumnComment(&content, idField.Comment)
			content.WriteString(g.getIdFieldSQL(idField))
			continue
		}

		content.WriteString(",\n")
		writeColumnComment(&content, field.Comment)
		sqlType := g.getSQLType(field.Type)

		content.WriteString(fmt.Sprintf("  %s %s", g.column(field.Name), sqlType))

		if field.Unique {
			content.WriteString(" UNIQUE")
		}

		// json defaults are applied in Go, mysql does not allow them on a JSON column
		if field.DefaultValue != nil && field.Type != schema.FieldTypeJSON {
			defaultVal := g.formatDefaultValue(field.DefaultValue, field.Type)
			content.WriteString(fmt.Sprintf(" DEFAULT %s", defaultVal))
		}

		if !field.Optional {
			content.WriteString(" NOT NULL")
		}

		// TODO write logic for DefaultFunc etc
	}

	// compound primary key from index.Primary, the parser clears the id flag so this
	// is the table's only PRIMARY KEY
	for _, idx := range entity.Indexes {
		if idx.Type != schema.IndexPrimary {
			continue
		}
		content.WriteString(",\n")
		content.WriteString(fmt.Sprintf("  PRIMARY KEY (%s)", strings.Join(g.indexColumns(idx), ", ")))
	}

	content.WriteString("\n);\n")

	content.WriteString(g.generateIndexSQL(entity))

	return content.String()
}

// generateIndexSQL emits CREATE INDEX for index.Fields. Primary keys are inline
// in the CREATE TABLE.
func (g *Generator) generateIndexSQL(entity schema.Entity) string {
	var content strings.Builder

	tableName := strings.ToLower(entity.Name)

	for _, idx := range entity.Indexes {
		if idx.Type != schema.IndexRegular {
			continue
		}

		name := idx.Name
		if name == "" {
			name = g.defaultIndexName(tableName, idx)
		}

		unique := ""
		if idx.Unique {
			unique = "UNIQUE "
		}

		content.WriteString(fmt.Sprintf("CREATE %sINDEX %s%s ON %s (%s);\n",
			unique,
			g.indexIfNotExists(),
			g.quote(name),
			g.quote(tableName),
			strings.Join(g.indexColumns(idx), ", "),
		))
	}

	return content.String()
}

// indexColumns renders each indexed column, appending DESC where asked. ASC is
// the SQL default and left implicit.
func (g *Generator) indexColumns(idx schema.Index) []string {
	cols := make([]string, len(idx.Columns))
	for i, c := range idx.Columns {
		if c.Desc {
			cols[i] = g.column(c.Name) + " DESC"
		} else {
			cols[i] = g.column(c.Name)
		}
	}
	return cols
}

// defaultIndexName builds the name from the table and columns, e.g. idx_user_env_is_active
func (g *Generator) defaultIndexName(tableName string, idx schema.Index) string {
	parts := append([]string{"idx", tableName}, idx.FieldNames()...)
	return strings.Join(parts, "_")
}

func (g *Generator) generateQueries(entities []schema.Entity, dir string) error {
	var content strings.Builder

	content.WriteString(util.GeneratedSQL)
	content.WriteString("-- SQLC compatible query definitions\n\n")

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

	var createQuery *schema.Query
	var createBulkQuery *schema.Query
	var updateQuery *schema.Query
	var deleteQuery *schema.Query
	var deleteAllQuery *schema.Query
	var getQueries []schema.Query
	var listQueries []schema.Query

	for _, query := range entity.SQLCQueries() {
		switch query.Type {
		case schema.QueryCreate:
			createQuery = &query
		case schema.QueryCreateBulk:
			createBulkQuery = &query
		case schema.QueryUpdate:
			updateQuery = &query
		case schema.QueryDelete:
			deleteQuery = &query
		case schema.QueryDeleteAll:
			deleteAllQuery = &query
		case schema.QueryGetBy:
			getQueries = append(getQueries, query)
		case schema.QueryListBy, schema.QueryListAll:
			listQueries = append(listQueries, query)
		}
	}

	content.WriteString(fmt.Sprintf("-- %s CRUD operations\n", entity.Name))

	// CREATE
	if createQuery != nil {
		g.writeInsertQuery(&content, entity, *createQuery, util.GenQueryName(*createQuery, entity.Name))
	}

	// CREATE BULK - a single-row insert; the sqlcWrap layer wraps it in a loop.
	if createBulkQuery != nil {
		g.writeInsertQuery(&content, entity, *createBulkQuery, util.GenQueryName(*createBulkQuery, entity.Name))
	}

	// READ (get by)
	for _, query := range getQueries {
		queryName := util.GenQueryName(query, entity.Name)
		writeQueryHeader(&content, query, queryName, "one")
		var whereParts []string
		for i, fieldName := range query.Fields {
			whereParts = append(whereParts, fmt.Sprintf("%s = %s", g.column(fieldName), g.getParameterPlaceholder(i+1)))
		}
		content.WriteString(fmt.Sprintf("SELECT * FROM %s WHERE %s;\n", g.quote(tableName), strings.Join(whereParts, " AND ")))
	}

	// LIST
	for _, query := range listQueries {
		queryName := util.GenQueryName(query, entity.Name)
		writeQueryHeader(&content, query, queryName, "many")
		var whereParts []string
		for _, fieldName := range query.Fields {
			whereParts = append(whereParts, fmt.Sprintf("%s = %s", g.column(fieldName), g.namedArg(fieldName)))
		}
		for _, filter := range query.Filters {
			switch filter.Type {
			case schema.QueryFilterEq:
				whereParts = append(whereParts, fmt.Sprintf("%s = %s", g.column(filter.Field), g.namedArg(filter.Field)))

			case schema.QueryFilterSearch:
				whereParts = append(whereParts, fmt.Sprintf("%s LIKE %s", g.column(filter.Field), g.namedArg(filter.Field)))

			case schema.QueryFilterRange:
				minArg := g.namedArg("min_" + filter.Field)
				maxArg := g.namedArg("max_" + filter.Field)
				whereParts = append(whereParts, fmt.Sprintf("%s BETWEEN %s AND %s", g.column(filter.Field), minArg, maxArg))
			}
		}
		selectSQL := fmt.Sprintf("SELECT * FROM %s", g.quote(tableName))
		// ListAll has no filters, so no WHERE clause
		if len(whereParts) > 0 {
			selectSQL += " WHERE " + strings.Join(whereParts, " AND ")
		}
		// pagination comes from the proto request, ListAll has no limit/offset there
		if query.Type == schema.QueryListBy {
			limitArg, offsetArg := g.limitOffsetArgs()
			selectSQL += fmt.Sprintf(" LIMIT %s OFFSET %s", limitArg, offsetArg)
		}
		content.WriteString(selectSQL + ";\n")
	}

	// UPDATE
	if updateQuery != nil {
		queryName := util.GenQueryName(*updateQuery, entity.Name)
		if g.supportsReturning() {
			writeQueryHeader(&content, *updateQuery, queryName, "one")
		} else {
			writeQueryHeader(&content, *updateQuery, queryName, "exec")
		}
		content.WriteString(fmt.Sprintf("UPDATE %s SET\n", g.quote(tableName)))

		var updateFields []string
		for _, field := range entity.Fields {
			canWrite := field.CanDbWrite()
			if !canWrite {
				continue
			}
			if field.IsID() || field.Immutable {
				continue
			}

			// For non-readable fields (like passwords), use COALESCE with nullable parameter
			canApiRead := entity.CanFieldRead(field)
			canApiWrite := entity.CanFieldWrite(field)
			acceptOptional := false
			if canApiWrite && (field.DefaultFunc != nil || field.DefaultValue != nil) {
				acceptOptional = true
			}
			var fieldUpdate string
			if !canApiRead || acceptOptional {
				// This makes the field optional in updates - if NULL is passed, keep existing value
				fieldUpdate = fmt.Sprintf("  %s = COALESCE(sqlc.narg('%s'), %s)", g.column(field.Name), field.Name, g.column(field.Name))
			} else {
				fieldUpdate = fmt.Sprintf("  %s = %s", g.column(field.Name), g.namedArg(field.Name))
			}
			updateFields = append(updateFields, fieldUpdate)
		}

		content.WriteString(strings.Join(updateFields, ",\n"))
		content.WriteString(fmt.Sprintf("\nWHERE %s = %s", g.column(idField.Name), g.namedArg(idField.Name)))
		if g.supportsReturning() {
			content.WriteString("\nRETURNING *;\n")
		} else {
			content.WriteString(";\n")
		}
	}

	// DELETE
	if deleteQuery != nil {
		writeQueryHeader(&content, *deleteQuery, util.GenQueryName(*deleteQuery, entity.Name), "exec")
		content.WriteString(fmt.Sprintf("DELETE FROM %s WHERE %s = %s;\n", g.quote(tableName), g.column(idField.Name), g.getParameterPlaceholder(1)))
	}

	// DELETE ALL
	if deleteAllQuery != nil {
		writeQueryHeader(&content, *deleteAllQuery, util.GenQueryName(*deleteAllQuery, entity.Name), "exec")
		content.WriteString(fmt.Sprintf("DELETE FROM %s;\n", g.quote(tableName)))
	}

	return content.String()
}

func (g *Generator) writeInsertQuery(content *strings.Builder, entity schema.Entity, query schema.Query, queryName string) {
	tableName := strings.ToLower(entity.Name)
	idField := entity.GetIdField()

	if g.supportsReturning() {
		writeQueryHeader(content, query, queryName, "one")
	} else {
		writeQueryHeader(content, query, queryName, "execlastid")
	}
	content.WriteString(fmt.Sprintf("INSERT INTO %s (\n", g.quote(tableName)))

	var insertFields []string
	var insertPlaceholders []string

	for _, field := range entity.Fields {
		canWrite := field.CanDbWrite()
		if !canWrite {
			continue
		}
		if field.IsID() && field.DefaultFunc == nil {
			continue
		}
		insertFields = append(insertFields, " "+g.column(field.Name))
		parameterPlaceholder := g.getParameterPlaceholder(len(insertPlaceholders) + 1)
		insertPlaceholders = append(insertPlaceholders, " "+parameterPlaceholder)
	}

	content.WriteString(fmt.Sprintf(" %s\n", strings.Join(insertFields, ",\n ")))
	content.WriteString(") VALUES (\n")
	content.WriteString(fmt.Sprintf(" %s\n", strings.Join(insertPlaceholders, ",\n ")))
	if g.supportsReturning() {
		content.WriteString(fmt.Sprintf(") RETURNING %s;\n", g.column(idField.Name)))
	} else {
		content.WriteString(");")
	}
}

func writeFile(filePath, content string) error {
	return os.WriteFile(filePath, []byte(content), 0644)
}
