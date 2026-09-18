package main

import (
	"fmt"
	"io/fs"
	"log"
	"net/http"

	"github.com/Zigelzi/roundtrip/cmd/web/view"
	"github.com/Zigelzi/roundtrip/internal/db"
)

// application holds the dependencies shared across HTTP handlers.
type application struct {
	queries *db.Queries
}

// staticFS is the embedded static directory, resolved once at startup. The
// sub-FS can only fail if the embed itself is broken (a build-time invariant),
// so a panic here surfaces a programmer error rather than a request-time one.
var staticFS = func() fs.FS {
	sub, err := fs.Sub(staticFiles, "static")
	if err != nil {
		panic(fmt.Sprintf("failed to load embedded static files: %v", err))
	}
	return sub
}()

// routes wires the HTTP handlers, including the embedded static file server
// that serves the Tailwind-built stylesheet at /static/tailwind.css.
func (app *application) routes() http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))
	mux.HandleFunc("/{$}", app.handleIndex)
	return mux
}

// handleIndex renders the landing page, greeting the user read from the database.
func (app *application) handleIndex(w http.ResponseWriter, r *http.Request) {
	user, err := app.queries.GetAppUser(r.Context())
	if err != nil {
		log.Printf("failed to get app user: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if err := view.Index(user.Name).Render(r.Context(), w); err != nil {
		log.Printf("failed to render index page: %v", err)
	}
}
