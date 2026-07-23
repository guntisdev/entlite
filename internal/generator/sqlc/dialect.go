package sqlc

import (
	"fmt"
	"strings"

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
	return fmt.Sprintf("  %s %s", field.Name, idType)
}

func (g *Generator) getIdFieldType(fieldType schema.FieldType, primary bool) string {
	// When an explicit index.Primary overrides the id field, it is no longer the
	// primary key, so emit a plain column type without the auto-increment & PRIMARY KEY clause.
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
	default:
		return "TEXT"
	}
}

func (g *Generator) getMySQLType(fieldType schema.FieldType) string {
	switch fieldType {
	case schema.FieldTypeString:
		return "TEXT"
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
	default:
		return "TEXT"
	}
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

func (g *Generator) namedArg(name string) string {
	return fmt.Sprintf(":%s", name)
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
