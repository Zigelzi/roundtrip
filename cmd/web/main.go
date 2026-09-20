package main

import (
	"context"
	"embed"
	"log"
	"net/http"
	"os"

	"github.com/Zigelzi/roundtrip/internal/db"
)

//go:embed static
var staticFiles embed.FS

const defaultDBPath = "roundtrip.db"

// defaultAddress only accepts connections from this machine. Set ADDR (e.g.
// ADDR=0.0.0.0:8080) to reach the app from a phone on the home network —
// there is no login, so only do that on a trusted network.
const defaultAddress = "127.0.0.1:8080"

func main() {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = defaultDBPath
	}

	address := os.Getenv("ADDR")
	if address == "" {
		address = defaultAddress
	}

	database, err := db.InitDB(dbPath)
	if err != nil {
		log.Fatalf("initializing database failed: %v", err)
	}
	defer database.Close()

	if err := db.RunMigrations(database); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	queries := db.New(database)

	// The real family names are not in the repository; a deployment supplies
	// them (see .env.example). Without FAMILY_NAMES the app still runs, with
	// the placeholder names from the migration.
	if err := db.ApplyFamilyNames(context.Background(), queries, os.Getenv("FAMILY_NAMES")); err != nil {
		log.Fatalf("failed to apply family names: %v", err)
	}

	app := newApplication(queries)

	log.Printf("listening on http://%s", address)
	if err := http.ListenAndServe(address, app.routes()); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}
