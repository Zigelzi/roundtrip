package main

import (
	"embed"
	"log"
	"net/http"
	"os"

	"github.com/Zigelzi/roundtrip/internal/db"
)

//go:embed static
var staticFiles embed.FS

const defaultDBPath = "roundtrip.db"
const address = "127.0.0.1:8080"

func main() {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = defaultDBPath
	}

	database, err := db.InitDB(dbPath)
	if err != nil {
		log.Fatalf("initializing database failed: %v", err)
	}
	defer database.Close()

	if err := db.RunMigrations(database); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	app := &application{queries: db.New(database)}

	log.Printf("listening on http://%s", address)
	if err := http.ListenAndServe(address, app.routes()); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}
