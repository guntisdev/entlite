package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	_ "modernc.org/sqlite"

	"github.com/guntisdev/entlite/examples/01-basic-entity/ent/gen/pb"
	"github.com/guntisdev/entlite/examples/01-basic-entity/server"
)

func main() {
	dbPath := "./db.db"

	database, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer database.Close()

	if err := initSchema(database); err != nil {
		log.Fatalf("Failed to initialize schema: %v", err)
	}

	userService := server.NewUserServiceServer(database)
	mux := http.NewServeMux()
	path, handler := pb.NewUserServiceHandler(userService)
	mux.Handle(path, handler)

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "OK")
	})

	fs := http.FileServer(http.Dir("./web/dist"))
	mux.Handle("/", fs)

	port := "8080"
	addr := fmt.Sprintf(":%s", port)
	log.Printf("Starting gRPC server on %s", addr)
	log.Printf("Web UI available at http://localhost%s", addr)
	log.Printf("Health check available at http://localhost%s/health", addr)

	if err := http.ListenAndServe(addr, h2c.NewHandler(mux, &http2.Server{})); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}

func initSchema(db *sql.DB) error {
	schemaBytes, err := os.ReadFile("ent/contract/sqlc/schema.sql")
	if err != nil {
		return fmt.Errorf("failed to read schema file: %w", err)
	}

	_, err = db.Exec(string(schemaBytes))
	if err != nil {
		return fmt.Errorf("failed to execute schema: %w", err)
	}

	return nil
}
