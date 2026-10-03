.PHONY: run build docs migrate-up migrate-down migrate-version migrate-force

# Default config path
CONFIG ?= config/config.yaml
# Migrations are embedded via ./migrations, so no golang-migrate CLI is needed.
MIGRATE := go run ./cmd/migrate -config "$(CONFIG)"


run:
	go run ./cmd/server/... -config "$(CONFIG)"

build:
	go build -o bin/server.exe ./cmd/server

docs:
	redocly build-docs api/openapi.yaml -o docs/api.html

migrate-up:
	$(MIGRATE) up

migrate-down:
	$(MIGRATE) down 1

migrate-version:
	$(MIGRATE) version

# Force the recorded version without running anything, e.g. make migrate-force V=47
migrate-force:
	$(MIGRATE) force $(V)