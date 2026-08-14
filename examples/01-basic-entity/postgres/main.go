package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"connectrpc.com/connect"
	"connectrpc.com/validate"
	_ "github.com/jackc/pgx/v5/stdlib"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	"github.com/guntisdev/entlite/examples/01-basic-entity/postgres/ent/gen/pb"
	"github.com/guntisdev/entlite/examples/01-basic-entity/postgres/server"
)

// Credentials match docker-compose.yml, see `make db`
const connStr = "postgres://postgres:postgres@localhost:5432/entlite?sslmode=disable"

func main() {
	port := flag.String("port", "8080", "http port to listen on")
	flag.Parse()

	database, err := sql.Open("pgx", connStr)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer database.Close()

	if err := waitForDatabase(database); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	if err := initSchema(database); err != nil {
		log.Fatalf("Failed to initialize schema: %v", err)
	}

	userService := server.NewUserServiceServer(database)
	mux := http.NewServeMux()
	path, handler := pb.NewUserServiceHandler(
		userService,
		connect.WithInterceptors(
			validate.NewInterceptor(),
			pb.NewValidateInterceptor(),
		),
	)
	mux.Handle(path, handler)

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "OK")
	})

	fs := http.FileServer(http.Dir("./web/dist"))
	mux.Handle("/", noStore(fs))

	addr := fmt.Sprintf(":%s", *port)
	log.Printf("Starting gRPC server on %s", addr)
	log.Printf("Web UI available at http://localhost%s", addr)
	log.Printf("Health check available at http://localhost%s/health", addr)

	if err := http.ListenAndServe(addr, h2c.NewHandler(mux, &http2.Server{})); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}

// noStore keeps the browser from caching the frontend, so switching between the
// examples on this port always serves the current one rather than a stale copy
func noStore(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		h.ServeHTTP(w, r)
	})
}

// waitForDatabase polls until postgres finishes its startup, which takes a few seconds
func waitForDatabase(db *sql.DB) error {
	deadline := time.Now().Add(60 * time.Second)
	var lastErr error

	for time.Now().Before(deadline) {
		if lastErr = db.Ping(); lastErr == nil {
			log.Println("Connected to database")
			return nil
		}
		time.Sleep(500 * time.Millisecond)
	}

	return fmt.Errorf("database not reachable after 60s, is it running? (`make db`): %w", lastErr)
}

func initSchema(db *sql.DB) error {
	schemaBytes, err := os.ReadFile("ent/contract/sqlc/schema.sql")
	if err != nil {
		return fmt.Errorf("failed to read schema file: %w", err)
	}

	applied := 0
	for _, stmt := range strings.Split(string(schemaBytes), ";") {
		if strings.TrimSpace(stripComments(stmt)) == "" {
			continue
		}

		if _, err := db.Exec(stmt); err != nil {
			log.Printf("Schema statement skipped: %v", err)
			continue
		}
		applied++
	}

	log.Printf("Database schema initialized (%d statements applied)", applied)
	return nil
}

func stripComments(stmt string) string {
	var b strings.Builder
	for _, line := range strings.Split(stmt, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "--") {
			continue
		}
		b.WriteString(line)
	}
	return b.String()
}
