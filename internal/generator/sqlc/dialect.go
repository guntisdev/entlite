package sqlc

import (
	"fmt"

	"github.com/guntisdev/entlite/internal/schema"
)

func (g *Generator) getIdentifierQuote() string {
	switch g.sqlDialect {
	case MySQL:
		return "`"
	case PostgreSQL, SQLite:
		return "\""
	}

	panic("unreachable: invalid SQL dialect")
}

// TODO pass actual id field from entity
func (g *Generator) getIdFieldSQL() string {
	switch g.sqlDialect {
	case PostgreSQL:
		return "  id SERIAL PRIMARY KEY"
	case SQLite:
		return "  id INTEGER PRIMARY KEY AUTOINCREMENT"
	case MySQL:
		return "  id INT AUTO_INCREMENT PRIMARY KEY"
	}

	panic("unreachable: invalid SQL dialect")
}

func (g *Generator) getSQLType(fieldType schema.FieldType) string {
	switch g.sqlDialect {
	case PostgreSQL:
		return g.getPostgresSQLType(fieldType)
	case SQLite:
		return g.getSQLiteType(fieldType)
	case MySQL:
		return g.getMySQLType(fieldType)
	}

	panic("unreachable: invalid SQL dialect")
}

func (g *Generator) getPostgresSQLType(fieldType schema.FieldType) string {
	switch fieldType {
	case schema.FieldTypeString:
		return "TEXT"
	case schema.FieldTypeInt32:
		return "INTEGER"
	case schema.FieldTypeBool:
		return "BOOLEAN"
	case schema.FieldTypeTime:
		return "TIMESTAMP WITH TIMEZONE"
	default:
		return "TEXT"
	}
}

func (g *Generator) getSQLiteType(fieldType schema.FieldType) string {
	switch fieldType {
	case schema.FieldTypeString:
		return "TEXT"
	case schema.FieldTypeInt32:
		return "INTEGER"
	case schema.FieldTypeBool:
		return "INTEGER" // sqlite uses integer for boolean
	case schema.FieldTypeTime:
		return "DATETIME"
	default:
		return "TEXT"
	}
}

func (g *Generator) getMySQLType(fieldType schema.FieldType) string {
	switch fieldType {
	case schema.FieldTypeString:
		return "TEXT"
	case schema.FieldTypeInt32:
		return "INTEGER"
	case schema.FieldTypeBool:
		return "TINYINT(1)"
	case schema.FieldTypeTime:
		return "TIMESTAMP"
	default:
		return "TEXT"
	}
}

func (g *Generator) formatDefaultValue(value interface{}, fieldType schema.FieldType) string {
	switch fieldType {
	case schema.FieldTypeBool:
		if b, ok := value.(bool); ok {
			if b {
				return "true"
			}
			return "false"
		}
		// case schema.FieldTypeTime:
	}

	return fmt.Sprintf("%v", value)
}

func (g *Generator) supportsReturning() bool {
	switch g.sqlDialect {
	case MySQL:
		return false
	case PostgreSQL, SQLite:
		return true
	}

	panic("unreachable: invalid SQL dialect")
}

func (g *Generator) getParameterPlaceholder(index int) string {
	switch g.sqlDialect {
	case PostgreSQL:
		return fmt.Sprintf("$%d", index)
	case SQLite:
		return "?"
	case MySQL:
		return "?"
	}

	panic("unreachable: invalid SQL dialect")
}
