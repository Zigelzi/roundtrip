# Roundtrip build commands.
#
# `make generate` runs all code generation: sqlc (SQL -> Go), templ
# (.templ -> Go), and the standalone Tailwind CLI (input CSS -> bundled
# stylesheet embedded by the web binary). `build`, `run`, and `test` depend
# on it, so a single command always produces a consistent app.

TAILWIND_IN  = ./cmd/web/tailwind.css
TAILWIND_OUT = ./cmd/web/static/tailwind.css

.PHONY: generate build run run-lan test dev dev/templ dev/server dev/tailwind

generate:
	sqlc generate
	templ generate
	tailwindcss -i $(TAILWIND_IN) -o $(TAILWIND_OUT) --minify

build: generate
	mkdir -p build
	go build -o ./build/roundtrip ./cmd/web

run: generate
	go run ./cmd/web

# run-lan makes the app reachable from phones on the home network. There is
# no login, so use it only on a trusted network. Under WSL this needs mirrored
# networking (see CLAUDE.md).
run-lan: generate
	@echo "On your phone, open http://$$(hostname -I | awk '{print $$1}'):8080"
	ADDR=0.0.0.0:8080 go run ./cmd/web

test: generate
	go test ./...

# dev runs three watchers in parallel so changes show up without restarting:
#   templ:    regenerates templates and runs a live-reload proxy on :8080
#             (reachable from a phone on the home network, like run-lan)
#   air:      rebuilds and restarts the app on :8081 when Go code or CSS changes
#   tailwind: rebuilds the CSS bundle when classes change
# Open http://<this machine>:8080 (or http://localhost:8080). After editing
# SQL, run `make generate` in another terminal: sqlc isn't watched.
dev:
	sqlc generate
	$(MAKE) -j3 dev/tailwind dev/server dev/templ

dev/templ:
	@echo "On your phone, open http://$$(hostname -I | awk '{print $$1}'):8080"
	templ generate -watch -proxy="http://127.0.0.1:8081" -proxyport=8080 -proxybind=0.0.0.0 -open-browser=false

dev/server:
	ADDR=127.0.0.1:8081 air

dev/tailwind:
	tailwindcss -i $(TAILWIND_IN) -o $(TAILWIND_OUT) --minify --watch=always
