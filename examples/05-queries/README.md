# 05-queries

One `Build` entity and a tour of every query the DSL generates, so you can see
the SQL, the proto and the Go method each one produces.

<!-- teaches:start -->
- Ready made queries: `Create()`, `CreateBulk()`, `Get()`, `Update()`, `Delete()`, `DeleteAll()`, `ListAll()`
- `GetBy("commit_sha")` and `ListBy("branch")` — the fields you name become `=` filters
- Filters spell out the operator: `Eq` is `=`, `Range` is a lower and upper bound, `Search` is `LIKE`
- `Asc()` and `Desc()` chained, the chain order is the order of the `ORDER BY` columns
- `Limit()`/`Offset()` take their values from the caller, a fixed `Limit(10)` stays in the SQL
- `Count()` returns the total before `LIMIT`, next to the page of rows
- `Distinct("branch")` returns column values, not rows, so the method gives `[]string`
- `Distinct("branch", "status")` dedupes on the tuple and returns the query's own row struct
- `GroupBy("branch")` with `Sum()` returns one row per group, in the query's own row struct
- `Sum()`, `Avg()`, `Min()`, `Max()` without a `GroupBy` fold the whole table into one row
- `Contracts(entlite.SQLC())` on a query keeps it out of the API, only the database method is generated
- A comment directly above a query becomes the comment of the generated SQL and rpc
<!-- teaches:end -->

## Entity

[`sqlite/ent/schema/build.go`](sqlite/ent/schema/build.go) is one `Build`, a
finished run of a CI pipeline. The fields exist to give the queries something to
work with:

| Field | Why it is here |
|---|---|
| `commit_sha` | `Unique()`, so it can be a `GetBy` key and an `Upsert` target |
| `branch`, `env`, `status` | repeat across rows, so `Distinct` has something to deduplicate |
| `author`, `message` | text to search with `LIKE` |
| `duration_ms` | a number to sort by, and what `Sum()`/`Avg()` fold |
| `started_at` | a timestamp for `Range` filters and `Desc()` sorting |
| `failed_tests` | `Optional()`, so it is a nullable column and a pointer in Go |

`Queries()` is grouped into blocks. Read it top to bottom, it is the point of
this example.

## What each block produces

**One row in.** `Create().Upsert("commit_sha")` puts `ON CONFLICT (commit_sha) DO
UPDATE` on the insert, so re-running the same commit overwrites the stored row
instead of failing on the unique column. `CreateBulk()` generates the same
single row insert and a wrapper that loops it.

**Filters.** A string field is shorthand for `filter.Eq`, so
`ListBy("branch", "status")` and `ListBy(filter.Eq("branch"), filter.Eq("status"))`
generate the same `WHERE branch = @branch AND status = @status`. Use the filter
form when you want another operator:

```sql
-- SearchBuilds
SELECT *, COUNT(*) OVER() AS total_size FROM "build"
WHERE env = @env
  AND started_at >= @min_started_at AND started_at <= @max_started_at
  AND message LIKE @message
  AND status = @status
ORDER BY started_at DESC LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');
```

One `filter.Range("started_at")` becomes two request parameters,
`min_started_at` and `max_started_at`.

**Limit.** `Limit()` asks the caller for the row count and puts a required
`limit` in the proto request. `Limit(10)` fixes it in the SQL instead, so
`ListLatestBuilds` has no `limit` in its request at all. `Offset()` needs a
`Limit()`.

**Count.** `Count()` adds `COUNT(*) OVER() AS total_size`, counted before
`LIMIT` cuts the page. The Go method returns it as a second value and the proto
response carries `total_size`, the name [AIP-158](https://google.aip.dev/158)
uses.

**Distinct.** One column returns the values themselves:

```sql
-- ListBranches
SELECT DISTINCT branch FROM "build" ORDER BY branch;
```

```go
branches, err := queries.ListBranches(ctx) // []string
```

```proto
message ListBranchesResponse {
  repeated string branch = 1;
}
```

Every column you name is part of the key, so `Distinct("branch", "status")`
dedupes on the pair and returns a row struct of those two columns, with a
`ListBranchStatusesRow` message to match. Sorting is limited to the selected
columns, and `Count()` does not combine with `Distinct()`, because the count is
evaluated before the rows are deduplicated.

**Aggregates.** `Sum()`, `Avg()`, `Min()` and `Max()` fold a column instead of
returning it. With a `GroupBy` you get one row per group:

```sql
-- BranchDurations
SELECT branch, CAST(COALESCE(SUM(duration_ms), 0) AS INTEGER) AS sum_duration_ms
FROM "build" GROUP BY branch ORDER BY branch;
```

```go
durations, err := queries.BranchDurations(ctx) // []BranchDurationsRow{Branch, SumDurationMs}
```

```proto
message BranchDurationsRow {
  string branch = 1;
  int64 sum_duration_ms = 2;
}
```

Without a `GroupBy` the whole table folds into one row, so `BuildTotals` returns
a single row and its response carries the values directly, with no `rows`.

The `CAST` and the `COALESCE` are not decoration. sqlc types an uncast aggregate
as `interface{}`, and it types a cast one as non-null, so an empty table or an
all-NULL group has to fold into a zero instead of failing the scan. That is why
`Max("failed_tests")` reads a nullable column and still returns a plain `int64`.

The rest of the rules: an aggregate query needs `Name()`, because the generated
name says nothing about what it folds. Integers widen to `int64` and `Avg()` is
always a `double`. Sorting takes a grouped column or an aggregate column, so
`Desc("sum_duration_ms")` gives the biggest group first and pairs with `Limit()`
for a top N. `Count()` does not combine with an aggregate, and `Min()`/`Max()` do
not take a time column — sqlite cannot cast a timestamp back without turning it
into a number.

**A query with no rpc.** `ListBuildsForCleanup` uses
`Contracts(entlite.SQLC())`, so it exists as a Go method for the server to call
and never reaches the proto contract.

## Not generated yet

The end of `Queries()` lists what is on the TODO list in the
[repository readme](../../README.md), commented out so you can see the shape
that is planned: a row count per group, `Having()` and `DeleteBy()`. Until then,
a query the DSL cannot express goes in a hand-written `.sql` file next to the
generated one — that is what [02-custom](../02-custom) is about.

## Run

This example has no web UI and no docker, only sqlite. It seeds six builds on
startup, so every query returns something.

```bash
cd sqlite
make run     # serves on :8080
```

```bash
# distinct branches
curl -X POST http://localhost:8080/proto.BuildService/ListBranches \
  -H 'Content-Type: application/json' -d '{}'
# {"branch":["feature-search", "hotfix-auth", "main"]}

# distinct branches of one env
curl -X POST http://localhost:8080/proto.BuildService/ListEnvBranches \
  -H 'Content-Type: application/json' -d '{"env":"staging"}'
# {"branch":["feature-search", "main"]}

# distinct pairs
curl -X POST http://localhost:8080/proto.BuildService/ListBranchStatuses \
  -H 'Content-Type: application/json' -d '{}'
# {"rows":[{"branch":"feature-search", "status":"passed"}, ...]}

# a page of one row, with the total next to it
curl -X POST http://localhost:8080/proto.BuildService/SearchBuilds \
  -H 'Content-Type: application/json' \
  -d '{"env":"prod","status":"failed","message":"%",
       "min_started_at":"2020-01-01T00:00:00Z",
       "max_started_at":"2030-01-01T00:00:00Z","limit":1,"offset":0}'
# {"rows":[{...}],"totalSize":"2"}

# the whole table folded into one row
curl -X POST http://localhost:8080/proto.BuildService/BuildTotals \
  -H 'Content-Type: application/json' -d '{}'
# {"sumDurationMs":"698000","avgDurationMs":116333.33333333333,"maxFailedTests":"11"}

# one row per branch
curl -X POST http://localhost:8080/proto.BuildService/BranchDurations \
  -H 'Content-Type: application/json' -d '{}'
# {"rows":[{"branch":"feature-search","sumDurationMs":"203000"}, ...]}

# the groups of one env, sorted and capped at three
curl -X POST http://localhost:8080/proto.BuildService/ListEnvBranchDurations \
  -H 'Content-Type: application/json' -d '{"env":"prod"}'
# {"rows":[{"branch":"hotfix-auth","sumDurationMs":"44000","avgDurationMs":44000}, ...]}
```

`message` is a `LIKE` pattern, so `%` matches everything and `release%` matches
one build.
