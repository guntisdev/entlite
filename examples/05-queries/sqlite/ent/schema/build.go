package schema

import (
	"github.com/guntisdev/entlite/pkg/entlite"
	"github.com/guntisdev/entlite/pkg/entlite/field"
	"github.com/guntisdev/entlite/pkg/entlite/filter"
	"github.com/guntisdev/entlite/pkg/entlite/query"
)

// Build is one run of a ci pipeline, recorded when it finishes
type Build struct {
	entlite.Schema
}

func (Build) Contracts() []entlite.Contract {
	return []entlite.Contract{
		entlite.SQLC(),
		entlite.PROTO(),
	}
}

func (Build) Fields() []entlite.Field {
	return []entlite.Field{
		// unique, so it can be a GetBy key and an Upsert target
		field.String("commit_sha").Unique(),
		// these repeat across rows, so Distinct has something to deduplicate
		field.String("branch"),
		field.String("env"),
		field.String("status"), // queued, running, passed or failed
		field.String("author"),
		field.String("message"),
		field.Int64("duration_ms"),
		field.Time("started_at"),
		field.Int("failed_tests").Optional(),
	}
}

// Queries is a tour of what the dsl generates. A comment directly above a query
// becomes the comment of the generated sql and rpc, the group headers below are
// separated by a blank line, so they stay in this file.
func (Build) Queries() []entlite.Query {
	return []entlite.Query{
		// ---- one row in ----

		// re-running the same commit overwrites the row instead of failing on the
		// unique column
		query.Create().Upsert("commit_sha"),
		// many rows in one call, the wrapper loops the single row insert
		query.CreateBulk(),

		// ---- one row out ----

		// Get() keys on the primary key
		query.Get(),
		// GetBy names the columns itself, they have to be unique together
		query.GetBy("commit_sha"),

		// ---- one row changed or gone ----

		query.Update(),
		query.Delete(),
		query.DeleteAll(),

		// ---- many rows out ----

		// no filters, the whole table
		query.ListAll(),
		// a string field becomes an = filter, the caller sends the value
		query.ListBy("branch"),
		// two fields are two = filters joined by AND, and the sort chain order is
		// the order of the columns in ORDER BY
		query.ListBy("branch", "status").
			Name("ListBranchStatusHistory").
			Desc("started_at").Asc("commit_sha"),
		// a fixed Limit(10) stays in the sql, so it is not part of the request
		query.ListBy("branch").
			Name("ListLatestBuilds").
			Desc("started_at").Limit(10),

		// ---- filters spell out the operator ----

		// Count() returns how many rows match, counted before Limit cuts the page.
		// Limit() and Offset() take their values from the caller.
		query.ListBy(
			filter.Eq("env"),           // env = :env
			filter.Range("started_at"), // :min <= started_at <= :max
			filter.Search("message"),   // message LIKE :message
			filter.Eq("status"),
		).Name("SearchBuilds").
			Desc("started_at").Count().Limit().Offset(),

		// ---- distinct columns instead of rows ----

		// one column returns the values, the method gives []string
		query.ListAll().Name("ListBranches").Distinct("branch").Asc("branch"),
		// the same, narrowed by a filter first
		query.ListBy(filter.Eq("env")).
			Name("ListEnvBranches").Distinct("branch").Asc("branch"),
		// several columns dedupe on the tuple and return a row struct
		query.ListAll().
			Name("ListBranchStatuses").Distinct("branch", "status").
			Asc("branch").Asc("status"),

		// ---- a query the server keeps to itself ----

		// query level Contracts() drops the rpc, only the database method is generated
		query.ListBy(filter.Range("started_at")).
			Name("ListBuildsForCleanup").Contracts(entlite.SQLC()),

		// ---- not generated yet, see the TODO list in the repository readme ----
		//
		// query.ListAll().Sum("duration_ms")      // one aggregate over the matched rows
		// query.ListAll().Avg("duration_ms")
		// query.ListAll().GroupBy("branch")       // one row per group
		// query.ListBy("branch").Having(...)      // filter the groups
		// query.DeleteBy("branch")                // delete by filter, not by key
	}
}
