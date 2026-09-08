package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"connectrpc.com/connect"
	"connectrpc.com/validate"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	_ "modernc.org/sqlite"

	"github.com/guntisdev/entlite/examples/05-queries/sqlite/ent/gen/db"
	"github.com/guntisdev/entlite/examples/05-queries/sqlite/ent/gen/pb"
	"github.com/guntisdev/entlite/examples/05-queries/sqlite/server"
)

func main() {
	port := flag.String("port", "8080", "http port to listen on")
	flag.Parse()

	dbPath := "./server/db.db"

	database, err := sql.Open("sqlite", dbPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer database.Close()

	if err := initSchema(database); err != nil {
		log.Fatalf("Failed to initialize schema: %v", err)
	}

	if err := seed(database); err != nil {
		log.Fatalf("Failed to seed builds: %v", err)
	}

	mux := http.NewServeMux()

	buildPath, buildHandler := pb.NewBuildServiceHandler(
		server.NewBuildServiceServer(database),
		connect.WithInterceptors(
			validate.NewInterceptor(),
			pb.NewValidateInterceptor(),
		),
	)
	mux.Handle(buildPath, buildHandler)

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "OK")
	})

	addr := fmt.Sprintf(":%s", *port)
	log.Printf("Starting gRPC server on %s", addr)
	log.Printf("Service: %s", buildPath)
	log.Printf("Health check available at http://localhost%s/health", addr)
	log.Printf("This example has no web ui, call the rpcs with curl, see the readme")

	if err := http.ListenAndServe(addr, h2c.NewHandler(mux, &http2.Server{})); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}

func initSchema(database *sql.DB) error {
	schemaBytes, err := os.ReadFile("ent/contract/sqlc/schema.sql")
	if err != nil {
		return fmt.Errorf("failed to read schema file: %w", err)
	}

	if _, err := database.Exec(string(schemaBytes)); err != nil {
		log.Printf("Schema initialization skipped (tables may already exist): %v", err)
		return nil
	}

	log.Println("Database schema initialized successfully")
	return nil
}

// seed writes a few builds, so every query has something to return. Upsert on
// commit_sha makes it safe to run again.
func seed(database *sql.DB) error {
	ctx := context.Background()
	queries := db.New(database)

	// spread the runs over the last day
	start := time.Now().UTC().Add(-24 * time.Hour)
	at := func(hours int) time.Time { return start.Add(time.Duration(hours) * time.Hour) }
	tests := func(count int32) *int32 { return &count }

	rows := []db.CreateBuildParams{
		{CommitSha: "a1b2c3d", Branch: "main", Env: "prod", Status: "passed",
			Author: "ada", Message: "release 1.4.0", DurationMs: 184_000, StartedAt: at(0), FailedTests: nil},
		{CommitSha: "b2c3d4e", Branch: "main", Env: "staging", Status: "passed",
			Author: "ada", Message: "bump deps", DurationMs: 171_000, StartedAt: at(1), FailedTests: nil},
		{CommitSha: "c3d4e5f", Branch: "main", Env: "prod", Status: "failed",
			Author: "linus", Message: "fix the cache key", DurationMs: 96_000, StartedAt: at(2), FailedTests: tests(3)},
		{CommitSha: "d4e5f6a", Branch: "feature-search", Env: "staging", Status: "passed",
			Author: "grace", Message: "add search filter", DurationMs: 203_000, StartedAt: at(3), FailedTests: nil},
		{CommitSha: "e5f6a7b", Branch: "feature-search", Env: "staging", Status: "running",
			Author: "grace", Message: "search pagination", DurationMs: 0, StartedAt: at(4), FailedTests: nil},
		{CommitSha: "f6a7b8c", Branch: "hotfix-auth", Env: "prod", Status: "failed",
			Author: "linus", Message: "token refresh", DurationMs: 44_000, StartedAt: at(5), FailedTests: tests(11)},
	}

	for _, row := range rows {
		if _, err := queries.CreateBuild(ctx, row); err != nil {
			return fmt.Errorf("failed to seed %s: %w", row.CommitSha, err)
		}
	}

	log.Printf("Seeded %d builds", len(rows))
	return nil
}
