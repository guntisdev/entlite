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

func (g *Generator) quote(str string) string {
	return g.getIdentifierQuote() + str + g.getIdentifierQuote()
}

func (g *Generator) getIdFieldSQL(field schema.Field) string {
	idType := g.getIdFieldType(field.Type)
	return fmt.Sprintf("  %s %s", field.Name, idType)
}

func (g *Generator) getIdFieldType(fieldType schema.FieldType) string {
	switch g.sqlDialect {
	case PostgreSQL:
		switch fieldType {
		case schema.FieldTypeInt32:
			return "SERIAL PRIMARY KEY"
		// case schema.FieldTypeInt64:
		//	return "BIGSERIAL PRIMARY KEY"
		case schema.FieldTypeString:
			return "TEXT PRIMARY KEY"
		default:
			return "SERIAL PRIMARY KEY"
		}
	case SQLite:
		switch fieldType {
		case schema.FieldTypeInt32:
			return "INTEGER PRIMARY KEY AUTOINCREMENT"
		// case schema.FieldTypeInt64:
		//	return "INTEGER PRIMARY KEY AUTOINCREMENT"
		case schema.FieldTypeString:
			return "TEXT PRIMARY KEY"
		default:
			return "INTEGER PRIMARY KEY AUTOINCREMENT"
		}
	case MySQL:
		switch fieldType {
		case schema.FieldTypeInt32:
			return "INT AUTO_INCREMENT PRIMARY KEY"
		// case schema.FieldTypeInt64:
		//	return "BIGINT AUTO_INCREMENT PRIMARY KEY"
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
