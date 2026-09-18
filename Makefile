# Roundtrip build commands.
#
# `make generate` runs all code generation: sqlc (SQL -> Go), templ
# (.templ -> Go), and the standalone Tailwind CLI (input CSS -> bundled
# stylesheet embedded by the web binary). `build`, `run`, and `test` depend
# on it, so a single command always produces a consistent app.

TAILWIND_IN  = ./cmd/web/tailwind.css
TAILWIND_OUT = ./cmd/web/static/tailwind.css

.PHONY: generate build run test

generate:
	sqlc generate
	templ generate
	tailwindcss -i $(TAILWIND_IN) -o $(TAILWIND_OUT) --minify

build: generate
	mkdir -p build
	go build -o ./build/roundtrip ./cmd/web

run: generate
	go run ./cmd/web

test: generate
	go test ./...
