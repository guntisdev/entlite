package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"connectrpc.com/connect"
	"connectrpc.com/validate"
	_ "github.com/jackc/pgx/v5/stdlib"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	"github.com/guntisdev/entlite/examples/02-postgres-entity/ent/gen/pb"
	"github.com/guntisdev/entlite/examples/02-postgres-entity/server"
)

const (
	containerName  = "entlite-postgres-02"
	postgresImage  = "postgres:17-alpine"
	postgresPort   = "5432"
	defaultConnStr = "postgres://postgres:postgres@localhost:" + postgresPort + "/entlite?sslmode=disable"
)

func main() {
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		// No external database configured, so run a throwaway one in Docker.
		if err := startPostgresContainer(); err != nil {
			log.Fatalf("Failed to start postgres container: %v", err)
		}
		connStr = defaultConnStr
	}

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

// startPostgresContainer reuses the example container or creates a new one
func startPostgresContainer() error {
	if _, err := exec.LookPath("docker"); err != nil {
		return fmt.Errorf("docker not found in PATH; start postgres yourself and set DATABASE_URL: %w", err)
	}

	state, err := exec.Command("docker", "inspect", "-f", "{{.State.Running}}", containerName).Output()
	switch {
	case err == nil && strings.TrimSpace(string(state)) == "true":
		log.Printf("Reusing running container %q", containerName)
		return nil
	case err == nil:
		log.Printf("Starting existing container %q", containerName)
		if out, err := exec.Command("docker", "start", containerName).CombinedOutput(); err != nil {
			return fmt.Errorf("docker start: %w: %s", err, out)
		}
		return nil
	}

	log.Printf("Creating container %q from %s", containerName, postgresImage)
	out, err := exec.Command("docker", "run", "-d",
		"--name", containerName,
		"-p", postgresPort+":5432",
		"-e", "POSTGRES_USER=postgres",
		"-e", "POSTGRES_PASSWORD=postgres",
		"-e", "POSTGRES_DB=entlite",
		postgresImage,
	).CombinedOutput()
	if err != nil {
		return fmt.Errorf("docker run: %w: %s", err, out)
	}

	return nil
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

	return fmt.Errorf("database not reachable after 60s: %w", lastErr)
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
