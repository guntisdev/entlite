package main

import (
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

	"github.com/guntisdev/entlite/examples/01-basic-entity/sqlite/ent/gen/pb"
	"github.com/guntisdev/entlite/examples/01-basic-entity/sqlite/server"
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
