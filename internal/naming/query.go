package naming

import "strings"

// a query name is a Go name, not a sql name
// sqlc copies it from the -- name: line into its method. the rpc uses the same name

// insert one row
func CreateQueryName(entity string) string { return "Create" + entity }

// insert many rows
func CreateBulkQueryName(entity string) string { return "CreateBulk" + entity }

// update by primary key
func UpdateQueryName(entity string) string { return "Update" + entity }

// delete by primary key
func DeleteQueryName(entity string) string { return "Delete" + entity }

// delete every row
func DeleteAllQueryName(entity string) string { return "DeleteAll" + entity }

// get one row by the given fields
func GetByQueryName(entity string, fields []string) string {
	return "Get" + entity + "By" + QueryWords(fields)
}

// list every row. distinct can be empty
func ListAllQueryName(entity string, distinct []string) string {
	return "ListAll" + entity + distinctWords(distinct)
}

// list some rows. distinct, fields and filters can be empty
func ListByQueryName(entity string, distinct, fields, filters []string) string {
	name := "List" + entity + distinctWords(distinct)

	if words := QueryWords(fields); words != "" {
		name += "By" + words
	}
	if words := QueryWords(filters); words != "" {
		name += "FilterBy" + words
	}

	return name
}

func AggregateColumn(fn, field string) string {
	return fn + "_" + field
}

func distinctWords(distinct []string) string {
	if len(distinct) == 0 {
		return ""
	}

	return "Distinct" + QueryWords(distinct)
}

// field names as the words of a query name: is_active -> IsActive
// no special words, so org_id -> OrgId
// a word starting with a digit stays joined, so field_2 -> Field2
// ProtocGoName does both of those differently, so this is its own rule
func QueryWords(fields []string) string {
	var b strings.Builder

	for _, field := range fields {
		for _, part := range strings.Split(field, "_") {
			if part == "" {
				continue
			}
			b.WriteString(strings.ToUpper(part[:1]))
			if len(part) > 1 {
				b.WriteString(part[1:])
			}
		}
	}

	return b.String()
}
