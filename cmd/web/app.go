package main

import (
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"time"
	_ "time/tzdata" // embed the zone database so Europe/Helsinki loads on a bare Pi

	"github.com/a-h/templ"

	"github.com/Zigelzi/roundtrip/cmd/web/view"
	"github.com/Zigelzi/roundtrip/internal/db"
)

// familyBucketID is the family_member row that means "the whole family", not
// a person. It is seeded out of range (see the migration comment in
// sql/schema/20260920130000_add_family_bucket.sql) so internal/db/config.go's
// positional FAMILY_NAMES mapping can never reach it. It lives here rather
// than beside one page because it spans the schema, both pages and the view
// layer. sql/queries/item.sql's ListPeople repeats the number as a literal,
// because sqlc cannot reach a Go constant; TestListPeopleExcludesTheFamilyBucket
// keeps the two in step.
const familyBucketID int64 = 100

// application holds the dependencies shared across HTTP handlers.
type application struct {
	queries *db.Queries
	// now is the clock; tests replace it to pin "today".
	now func() time.Time
}

func newApplication(queries *db.Queries) *application {
	return &application{queries: queries, now: time.Now}
}

// helsinki is the family's time zone; "today" means the date in Finland.
var helsinki = func() *time.Location {
	loc, err := time.LoadLocation("Europe/Helsinki")
	if err != nil {
		panic(fmt.Sprintf("failed to load Europe/Helsinki: %v", err))
	}
	return loc
}()

// today returns the current calendar date in Finland, as midnight UTC so it
// compares directly with dates parsed from the database or forms.
func (app *application) today() time.Time {
	y, m, d := app.now().In(helsinki).Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
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
	mux.HandleFunc("GET /{$}", app.handleIndex)
	mux.HandleFunc("GET /trips/new", app.handleNewTrip)
	mux.HandleFunc("POST /trips", app.handleCreateTrip)
	mux.HandleFunc("GET /trips/{id}", app.handleTrip)
	mux.HandleFunc("GET /trips/{id}/pack", app.handlePackPage)
	mux.HandleFunc("POST /trips/{id}/members/{memberID}/items/{itemID}/pack", app.handlePackItem)
	mux.HandleFunc("POST /trips/{id}/members/{memberID}/items/{itemID}/unpack", app.handleUnpackItem)
	mux.HandleFunc("POST /trips/{id}/members/{memberID}/items", app.handleAddItem)
	mux.HandleFunc("POST /trips/{id}/items/{itemID}/delete", app.handleRemoveItem)
	return skipDevReloadForHTMX(mux)
}

// handleIndex renders the front page: the family's trips.
func (app *application) handleIndex(w http.ResponseWriter, r *http.Request) {
	rows, err := app.queries.ListTrips(r.Context())
	if err != nil {
		app.serverError(w, "list trips", err)
		return
	}
	trips, err := sortTrips(rows, app.today())
	if err != nil {
		app.serverError(w, "sort trips", err)
		return
	}
	app.render(w, r, http.StatusOK, view.Index(trips))
}

// skipDevReloadForHTMX marks htmx replies so templ's live-reload proxy
// (make dev) leaves them alone: it would otherwise wrap each fragment in
// <html><body> and add a reload script to it. The header does nothing outside
// development.
func skipDevReloadForHTMX(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if isHTMX(r) {
			w.Header().Set("templ-skip-modify", "true")
		}
		next.ServeHTTP(w, r)
	})
}

// render writes a templ component with the given status code.
func (app *application) render(w http.ResponseWriter, r *http.Request, status int, c templ.Component) {
	w.WriteHeader(status)
	if err := c.Render(r.Context(), w); err != nil {
		log.Printf("failed to render page: %v", err)
	}
}

func (app *application) serverError(w http.ResponseWriter, action string, err error) {
	log.Printf("failed to %s: %v", action, err)
	http.Error(w, "internal error", http.StatusInternalServerError)
}
