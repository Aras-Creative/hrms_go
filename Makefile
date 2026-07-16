.PHONY: run build docs migrate-up

# Default config path
CONFIG ?= config/config.yaml
DB_URL := postgres://postgres:aras@localhost:5432/sekantor_v2_db?sslmode=disable


run:
	go run ./cmd/server/... -config "$(CONFIG)"

build:
	go build -o bin/server.exe ./cmd/server

docs:
	redocly build-docs api/openapi.yaml -o docs/api.html

migrate-up:
	migrate -path migrations -database "$(DB_URL)" up

migrate-down:
	migrate -path migrations -database "$(DB_URL)" down 1