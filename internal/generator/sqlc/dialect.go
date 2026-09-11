package sqlc

import (
	"fmt"
	"strings"

	"github.com/guntisdev/entlite/internal/naming"
	"github.com/guntisdev/entlite/internal/schema"
)

func (g *Generator) getIdentifierQuote() string {
	switch g.sqlDialect {
	case schema.MySQL:
		return "`"
	case schema.PostgreSQL, schema.SQLite:
		return "\""
	}

	panic("unreachable: invalid SQL dialect")
}

func (g *Generator) quote(str string) string {
	return g.getIdentifierQuote() + str + g.getIdentifierQuote()
}

func (g *Generator) getIdFieldSQL(field schema.Field) string {
	idType := g.getIdFieldType(field.Type, field.Primary)
	// index.Primary has the key. id is a plain column now, so it needs NOT NULL
	if !field.Primary && !field.Optional {
		return fmt.Sprintf("  %s %s NOT NULL", g.column(field.Name), idType)
	}

	return fmt.Sprintf("  %s %s", g.column(field.Name), idType)
}

func (g *Generator) getIdFieldType(fieldType schema.FieldType, primary bool) string {
	// index.Primary overrides the id field, it is no longer primary key
	if !primary {
		return g.getSQLType(fieldType)
	}

	switch g.sqlDialect {
	case schema.PostgreSQL:
		switch fieldType {
		case schema.FieldTypeInt:
			return "SERIAL PRIMARY KEY"
		case schema.FieldTypeInt64:
			return "BIGSERIAL PRIMARY KEY"
		case schema.FieldTypeString:
			return "TEXT PRIMARY KEY"
		default:
			return "SERIAL PRIMARY KEY"
		}
	case schema.SQLite:
		switch fieldType {
		case schema.FieldTypeInt:
			return "INTEGER PRIMARY KEY AUTOINCREMENT"
		case schema.FieldTypeInt64:
			return "INTEGER PRIMARY KEY AUTOINCREMENT"
		case schema.FieldTypeString:
			return "TEXT PRIMARY KEY"
		default:
			return "INTEGER PRIMARY KEY AUTOINCREMENT"
		}
	case schema.MySQL:
		switch fieldType {
		case schema.FieldTypeInt:
			return "INT AUTO_INCREMENT PRIMARY KEY"
		case schema.FieldTypeInt64:
			return "BIGINT AUTO_INCREMENT PRIMARY KEY"
		case schema.FieldTypeString:
			return "VARCHAR(36) PRIMARY KEY" // UUID or ULID or similar string ID
		default:
			return "INT AUTO_INCREMENT PRIMARY KEY"
		}
	}

	panic("unreachable: invalid SQL dialect")
}

// returns the column type used for a field in the given dialect.
func SQLTypeFor(dialect schema.SQLDialect, fieldType schema.FieldType) string {
	return NewGenerator(dialect).getSQLType(fieldType)
}

func (g *Generator) getSQLType(fieldType schema.FieldType) string {
	switch g.sqlDialect {
	case schema.PostgreSQL:
		return g.getPostgresSQLType(fieldType)
	case schema.SQLite:
		return g.getSQLiteType(fieldType)
	case schema.MySQL:
		return g.getMySQLType(fieldType)
	}

	panic("unreachable: invalid SQL dialect")
}

func (g *Generator) getPostgresSQLType(fieldType schema.FieldType) string {
	switch fieldType {
	case schema.FieldTypeString:
		return "TEXT"
	case schema.FieldTypeInt:
		return "INT"
	case schema.FieldTypeInt64:
		return "BIGINT"
	case schema.FieldTypeFloat:
		return "DOUBLE PRECISION"
	case schema.FieldTypeBool:
		return "BOOLEAN"
	case schema.FieldTypeTime:
		return "TIMESTAMPTZ"
	case schema.FieldTypeByte:
		return "BYTEA"
	case schema.FieldTypeJSON:
		return "JSONB"
	default:
		return "TEXT"
	}
}

func (g *Generator) getSQLiteType(fieldType schema.FieldType) string {
	switch fieldType {
	case schema.FieldTypeString:
		return "TEXT"
	case schema.FieldTypeInt:
		return "INTEGER"
	case schema.FieldTypeInt64:
		return "INTEGER"
	case schema.FieldTypeFloat:
		return "REAL"
	case schema.FieldTypeBool:
		return "INTEGER" // sqlite uses integer for boolean
	case schema.FieldTypeTime:
		return "DATETIME"
	case schema.FieldTypeByte:
		return "BLOB"
	case schema.FieldTypeJSON:
		return "TEXT" // sqlite has no json type, generateTableSQL adds a json_valid CHECK
	default:
		return "TEXT"
	}
}

func (g *Generator) getMySQLType(fieldType schema.FieldType) string {
	switch fieldType {
	case schema.FieldTypeString:
		// VARCHAR not TEXT: MySQL cannot index TEXT without a key length, which breaks Unique()
		return "VARCHAR(255)"
	case schema.FieldTypeInt:
		return "INT"
	case schema.FieldTypeInt64:
		return "BIGINT"
	case schema.FieldTypeFloat:
		return "DOUBLE"
	case schema.FieldTypeBool:
		return "TINYINT(1)"
	case schema.FieldTypeTime:
		return "TIMESTAMP"
	case schema.FieldTypeByte:
		return "BLOB"
	case schema.FieldTypeJSON:
		return "JSON"
	default:
		return "TEXT"
	}
}

// returns the column constraint that keeps invalid json out
func (g *Generator) jsonCheck(field schema.Field) string {
	if field.Type != schema.FieldTypeJSON {
		return ""
	}
	usesNativeJSONColumn := g.sqlDialect == schema.PostgreSQL ||
		(g.sqlDialect == schema.MySQL && !field.Optional)
	if usesNativeJSONColumn {
		return ""
	}

	return fmt.Sprintf("CHECK (json_valid(%s))", g.column(field.Name))
}

func (g *Generator) formatDefaultValue(value any, fieldType schema.FieldType) string {
	switch fieldType {
	case schema.FieldTypeBool:
		if b, ok := value.(bool); ok {
			if b {
				return "true"
			}
			return "false"
		}
	case schema.FieldTypeString:
		// String literals must be single-quoted; escape embedded quotes.
		s := fmt.Sprintf("%v", value)
		return "'" + strings.ReplaceAll(s, "'", "''") + "'"
	}

	return fmt.Sprintf("%v", value)
}

func (g *Generator) supportsReturning() bool {
	switch g.sqlDialect {
	case schema.MySQL:
		return false
	case schema.PostgreSQL, schema.SQLite:
		return true
	}

	panic("unreachable: invalid SQL dialect")
}

// Postgres and sqlite name target, mysql cannot name one and fires on every unique key of the table.
func (g *Generator) upsertClause(entity schema.Entity, query schema.Query) string {
	target := entity.UpsertTarget(query)

	switch g.sqlDialect {
	case schema.MySQL:
		if query.UpsertIgnore {
			// a no-op assignment keeps the stored row, INSERT IGNORE would hide every other error too
			column := g.column(target[0])
			return fmt.Sprintf("ON DUPLICATE KEY UPDATE %s = %s", column, column)
		}
		return "ON DUPLICATE KEY UPDATE\n" + strings.Join(g.upsertSetColumns(entity, target), ",\n")

	case schema.PostgreSQL, schema.SQLite:
		columns := make([]string, len(target))
		for i, name := range target {
			columns[i] = g.column(name)
		}
		clause := fmt.Sprintf("ON CONFLICT (%s)", strings.Join(columns, ", "))
		if query.UpsertIgnore {
			return clause + " DO NOTHING"
		}
		return clause + " DO UPDATE SET\n" + strings.Join(g.upsertSetColumns(entity, target), ",\n")
	}

	panic("unreachable: invalid SQL dialect")
}

func (g *Generator) upsertSetColumns(entity schema.Entity, target []string) []string {
	fields := entity.UpsertSetFields(target)
	columns := make([]string, 0, len(fields))

	for _, field := range fields {
		columns = append(columns, fmt.Sprintf("  %s = %s", g.column(field.Name), g.upsertValue(field.Name)))
	}

	return columns
}

func (g *Generator) upsertValue(fieldName string) string {
	switch g.sqlDialect {
	case schema.MySQL:
		// sqlc's mysql parser does not take the 8.0.19 row alias, VALUES() works everywhere
		return fmt.Sprintf("VALUES(%s)", g.column(fieldName))
	case schema.PostgreSQL, schema.SQLite:
		return fmt.Sprintf("excluded.%s", g.column(fieldName))
	}

	panic("unreachable: invalid SQL dialect")
}

// reports warning for mysql
func (g *Generator) upsertWarning(entity schema.Entity, query schema.Query) string {
	if g.sqlDialect != schema.MySQL || !query.Upsert {
		return ""
	}

	target := entity.UpsertTarget(query)
	others := entity.OtherCollidableConstraints(target)
	if len(others) == 0 {
		return ""
	}

	rendered := make([]string, len(others))
	for i, columns := range others {
		rendered[i] = "(" + strings.Join(columns, ", ") + ")"
	}

	return fmt.Sprintf("mysql ON DUPLICATE KEY UPDATE takes no conflict target, so it ignores Upsert(%s) and fires on %s as well",
		strings.Join(target, ", "),
		strings.Join(rendered, " and "),
	)
}

func (g *Generator) indexIfNotExists() string {
	switch g.sqlDialect {
	case schema.MySQL:
		return ""
	case schema.PostgreSQL, schema.SQLite:
		return "IF NOT EXISTS "
	}

	panic("unreachable: invalid SQL dialect")
}

func (g *Generator) limitOffsetArgs() (limit, offset string) {
	switch g.sqlDialect {
	case schema.MySQL:
		return "?", "?"
	case schema.PostgreSQL, schema.SQLite:
		// limit and offset are reserved, so namedArg gives the sqlc.arg() form
		return g.namedArg("limit"), g.namedArg("offset")
	}

	panic("unreachable: invalid SQL dialect")
}

// COALESCE folds an empty table or an all null group into a zero instead of a failed scan
func (g *Generator) aggregateExpr(entity schema.Entity, aggregate schema.Aggregate) string {
	expr := fmt.Sprintf("%s(%s)", strings.ToUpper(string(aggregate.Func)), g.column(aggregate.Field))
	alias := g.column(naming.AggregateColumn(string(aggregate.Func), aggregate.Field))

	field, found := entity.GetFieldByName(aggregate.Field)
	if !found {
		return fmt.Sprintf("%s AS %s", expr, alias)
	}

	resultType := schema.AggregateResultType(aggregate.Func, field.Type)
	zero := "0"
	if resultType == schema.FieldTypeString {
		zero = "''"
	}
	expr = fmt.Sprintf("COALESCE(%s, %s)", expr, zero)

	return fmt.Sprintf("%s AS %s", g.aggregateTyped(expr, resultType), alias)
}

// mysql only types a string aggregate through CONCAT, a CAST leaves it interface{}
func (g *Generator) aggregateTyped(expr string, resultType schema.FieldType) string {
	if resultType == schema.FieldTypeString && g.sqlDialect == schema.MySQL {
		return fmt.Sprintf("CONCAT(%s)", expr)
	}

	return fmt.Sprintf("CAST(%s AS %s)", expr, g.aggregateCastType(resultType))
}

func (g *Generator) aggregateCastType(resultType schema.FieldType) string {
	switch resultType {
	case schema.FieldTypeFloat:
		switch g.sqlDialect {
		case schema.MySQL:
			return "DOUBLE"
		case schema.PostgreSQL:
			return "DOUBLE PRECISION"
		case schema.SQLite:
			return "REAL"
		}
	case schema.FieldTypeString:
		switch g.sqlDialect {
		case schema.PostgreSQL, schema.SQLite:
			return "TEXT"
		}
	default:
		switch g.sqlDialect {
		case schema.MySQL:
			return "SIGNED"
		case schema.PostgreSQL:
			return "BIGINT"
		case schema.SQLite:
			return "INTEGER"
		}
	}

	panic("unreachable: invalid SQL dialect")
}

// COUNT(*) OVER() counts every matching row, LIMIT cuts the page afterwards
func (g *Generator) totalSizeExpr() string {
	switch g.sqlDialect {
	case schema.MySQL:
		// mysql types a bare window count as unknown, the cast keeps it an int64
		return "CAST(COUNT(*) OVER() AS SIGNED)"
	case schema.PostgreSQL, schema.SQLite:
		return "COUNT(*) OVER()"
	}

	panic("unreachable: invalid SQL dialect")
}

func (g *Generator) namedArg(name string) string {
	switch g.sqlDialect {
	case schema.MySQL:
		// sqlc's MySQL parser reads @name as a user variable and drops it, so use sqlc.arg()
		return fmt.Sprintf("sqlc.arg('%s')", name)
	case schema.PostgreSQL, schema.SQLite:
		// postgres parses @limit as the keyword, sqlc.arg() takes any name
		if isReservedWord(name) {
			return fmt.Sprintf("sqlc.arg('%s')", name)
		}
		return fmt.Sprintf("@%s", name)
	}

	panic("unreachable: invalid SQL dialect")
}

// sqlc.narg() declares a nullable named arg regardless of dialect or the column's
// own nullability - passing NULL is how a caller skips an optional filter/update field.
func (g *Generator) nargArg(name string) string {
	return fmt.Sprintf("sqlc.narg('%s')", name)
}

func (g *Generator) getParameterPlaceholder(index int) string {
	switch g.sqlDialect {
	case schema.PostgreSQL:
		return fmt.Sprintf("$%d", index)
	case schema.SQLite:
		return "?"
	case schema.MySQL:
		return "?"
	}

	panic("unreachable: invalid SQL dialect")
}

// BETWEEN on sqlite loses its min_/max_ args unless the range is the first filter, >= / <= always binds
func (g *Generator) rangeClause(field string) string {
	column := g.column(field)
	minArg := g.namedArg("min_" + field)
	maxArg := g.namedArg("max_" + field)

	switch g.sqlDialect {
	case schema.MySQL, schema.PostgreSQL:
		return fmt.Sprintf("%s BETWEEN %s AND %s", column, minArg, maxArg)
	case schema.SQLite:
		return fmt.Sprintf("%s >= %s AND %s <= %s", column, minArg, column, maxArg)
	}

	panic("unreachable: invalid SQL dialect")
}

// each bound is skipped independently when its arg is NULL, so BETWEEN doesn't
// apply here - it can't express "no lower bound" or "no upper bound"
func (g *Generator) rangeClauseOptional(field string) string {
	column := g.column(field)
	minArg := g.nargArg("min_" + field)
	maxArg := g.nargArg("max_" + field)

	return fmt.Sprintf("(%s IS NULL OR %s >= %s) AND (%s IS NULL OR %s <= %s)", minArg, column, minArg, maxArg, column, maxArg)
}
