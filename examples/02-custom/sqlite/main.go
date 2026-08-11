package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"connectrpc.com/connect"
	"connectrpc.com/validate"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	_ "modernc.org/sqlite"

	"github.com/guntisdev/entlite/examples/02-custom/sqlite/ent/gen/pb"
	"github.com/guntisdev/entlite/examples/02-custom/sqlite/server"
)

func main() {
	dbPath := "./server/db.db"

	database, err := sql.Open("sqlite", dbPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer database.Close()

	if err := initSchema(database); err != nil {
		log.Fatalf("Failed to initialize schema: %v", err)
	}

	mux := http.NewServeMux()

	interceptors := connect.WithInterceptors(
		validate.NewInterceptor(),
		pb.NewValidateInterceptor(),
	)

	sensorPath, sensorHandler := pb.NewSensorServiceHandler(
		server.NewSensorServiceServer(database),
		interceptors,
	)
	mux.Handle(sensorPath, sensorHandler)

	readingPath, readingHandler := pb.NewReadingServiceHandler(
		server.NewReadingServiceServer(database),
		interceptors,
	)
	mux.Handle(readingPath, readingHandler)

	analyticsPath, analyticsHandler := pb.NewSensorAnalyticsServiceHandler(
		server.NewSensorAnalyticsServiceServer(database),
		interceptors,
	)
	mux.Handle(analyticsPath, analyticsHandler)

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "OK")
	})

	port := "8080"
	addr := fmt.Sprintf(":%s", port)
	log.Printf("Starting gRPC server on %s", addr)
	log.Printf("Services: %s %s %s", sensorPath, readingPath, analyticsPath)
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
		log.Printf("Schema initialization skipped (tables may already exist): %v", err)
		return nil
	}

	log.Println("Database schema initialized successfully")
	return nil
}
