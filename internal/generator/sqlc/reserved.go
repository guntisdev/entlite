package sqlc

import "strings"

// Words reserved by at least one supported dialect
var reservedWords = newWordSet(
	"abort", "accessible", "action", "add", "all", "alter",
	"always", "analyse", "analyze", "and", "any", "array",
	"as", "asc", "asymmetric", "attach", "authorization", "autoincrement",
	"before", "between", "bigint", "binary", "blob", "both",
	"by", "call", "cascade", "case", "cast", "change",
	"char", "character", "check", "collate", "collation", "column",
	"concurrently", "condition", "conflict", "constraint", "continue", "convert",
	"create", "cross", "cube", "cume_dist", "current_catalog", "current_date",
	"current_role", "current_schema", "current_time", "current_timestamp", "current_user", "cursor",
	"database", "databases", "day_hour", "day_microsecond", "day_minute", "day_second",
	"dec", "decimal", "declare", "default", "deferrable", "deferred",
	"delayed", "delete", "dense_rank", "desc", "describe", "detach",
	"deterministic", "distinct", "distinctrow", "div", "do", "double",
	"drop", "dual", "each", "else", "elseif", "empty",
	"enclosed", "end", "escaped", "except", "exclude", "exclusive",
	"exists", "exit", "explain", "fail", "false", "fetch",
	"filter", "first_value", "float", "float4", "float8", "following",
	"for", "force", "foreign", "freeze", "from", "full",
	"fulltext", "function", "generated", "get", "glob", "grant",
	"group", "groups", "having", "high_priority", "hour_microsecond", "hour_minute",
	"hour_second", "if", "ignore", "ilike", "immediate", "in",
	"index", "indexed", "infile", "initially", "inner", "inout",
	"insensitive", "insert", "instead", "int", "int1", "int2",
	"int3", "int4", "int8", "integer", "intersect", "interval",
	"into", "io_after_gtids", "io_before_gtids", "is", "isnull", "iterate",
	"join", "json_table", "key", "keys", "kill", "lag",
	"last_value", "lateral", "lead", "leading", "leave", "left",
	"like", "limit", "linear", "lines", "load", "localtime",
	"localtimestamp", "lock", "long", "longblob", "longtext", "loop",
	"low_priority", "master_bind", "match", "maxvalue", "mediumblob", "mediumint",
	"mediumtext", "middleint", "minute_microsecond", "minute_second", "mod", "modifies",
	"natural", "no_write_to_binlog", "not", "nothing", "notnull", "nth_value",
	"ntile", "null", "nulls", "numeric", "offset", "on",
	"only", "optimize", "option", "optionally", "or", "order",
	"others", "out", "outer", "outfile", "over", "overlaps",
	"partition", "percent_rank", "placing", "plan", "pragma", "preceding",
	"precision", "primary", "procedure", "purge", "query", "raise",
	"range", "rank", "read", "reads", "real", "recursive",
	"references", "regexp", "reindex", "release", "rename", "repeat",
	"replace", "require", "resignal", "restrict", "returning", "revoke",
	"right", "rlike", "row", "row_number", "rows", "savepoint",
	"schema", "schemas", "second_microsecond", "select", "sensitive", "separator",
	"session_user", "set", "show", "signal", "similar", "some",
	"spatial", "specific", "sql", "sql_big_result", "sql_calc_found_rows", "sql_small_result",
	"sqlexception", "sqlstate", "sqlwarning", "ssl", "starting", "stored",
	"straight_join", "symmetric", "system", "table", "tablesample", "temp",
	"temporary", "terminated", "then", "ties", "tinyblob", "tinyint",
	"tinytext", "to", "trailing", "transaction", "trigger", "true",
	"unbounded", "undo", "union", "unique", "unlock", "unsigned",
	"update", "usage", "use", "user", "using", "utc_date",
	"utc_time", "utc_timestamp", "vacuum", "values", "varbinary", "varchar",
	"varcharacter", "variadic", "varying", "verbose", "view", "virtual",
	"when", "where", "while", "window", "with", "without",
	"write", "xor", "year_month", "zerofill",
)

func newWordSet(words ...string) map[string]bool {
	set := make(map[string]bool, len(words))
	for _, word := range words {
		set[word] = true
	}
	return set
}

func isReservedWord(name string) bool {
	return reservedWords[strings.ToLower(name)]
}

func (g *Generator) column(name string) string {
	if isReservedWord(name) {
		return g.quote(name)
	}

	return name
}
