package schema

type SQLDialect string

const (
	MySQL      SQLDialect = "mysql"
	SQLite     SQLDialect = "sqlite"
	PostgreSQL SQLDialect = "postgresql"
)
