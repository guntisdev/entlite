package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"connectrpc.com/connect"
	"connectrpc.com/validate"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	_ "modernc.org/sqlite"

	"github.com/guntisdev/entlite/examples/04-contracts/sqlite/ent/gen/pb"
	"github.com/guntisdev/entlite/examples/04-contracts/sqlite/server"
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

	if err := server.SeedRoster(context.Background(), database); err != nil {
		log.Fatalf("Failed to seed roster: %v", err)
	}

	mux := http.NewServeMux()

	interceptors := connect.WithInterceptors(
		validate.NewInterceptor(),
		pb.NewValidateInterceptor(),
	)

	// Match has both contracts, its handlers read and write the match table
	matchPath, matchHandler := pb.NewMatchServiceHandler(
		server.NewMatchServiceServer(database),
		interceptors,
	)
	mux.Handle(matchPath, matchHandler)

	// Standing has only the proto contract, its handler counts matches instead
	standingPath, standingHandler := pb.NewStandingServiceHandler(
		server.NewStandingServiceServer(database),
		interceptors,
	)
	mux.Handle(standingPath, standingHandler)

	// Player has both contracts, but proto is read only: the roster is
	// written by SeedRoster on the server and clients can only read it
	playerPath, playerHandler := pb.NewPlayerServiceHandler(
		server.NewPlayerServiceServer(database),
		interceptors,
	)
	mux.Handle(playerPath, playerHandler)

	// AuditEntry has only the sqlc contract, so it has no service at all.
	// This plain endpoint proves the rows exist without exposing them.
	mux.HandleFunc("/audit-count", func(w http.ResponseWriter, r *http.Request) {
		count, err := server.AuditCount(r.Context(), database)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Fprintf(w, "%d", count)
	})

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "OK")
	})

	fs := http.FileServer(http.Dir("./web/dist"))
	mux.Handle("/", noStore(fs))

	addr := fmt.Sprintf(":%s", *port)
	log.Printf("Starting gRPC server on %s", addr)
	log.Printf("Services: %s %s %s", matchPath, standingPath, playerPath)
	log.Printf("Web UI available at http://localhost%s", addr)
	log.Printf("Health check available at http://localhost%s/health", addr)

	if err := http.ListenAndServe(addr, h2c.NewHandler(mux, &http2.Server{})); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}

// noStore disables browser cache, so the port serves the current example
func noStore(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		h.ServeHTTP(w, r)
	})
}

func initSchema(db *sql.DB) error {
	schemaBytes, err := os.ReadFile("ent/contract/sqlc/schema.sql")
	if err != nil {
		return fmt.Errorf("failed to read schema file: %w", err)
	}

	_, err = db.Exec(string(schemaBytes))
	if err != nil {
		log.Printf("Schema initialization skipped (tables may already exist): %v", err)
		return nil
	}

	log.Println("Database schema initialized successfully")
	return nil
}
