-include .env
export

MIGRATE ?= migrate
MIGRATIONS_DIR := migrations

.PHONY: run dev build test lint tidy migrate-up migrate-down migrate-new seed

run:
	go run ./cmd/api

dev:
	air

build:
	go build -o bin/api ./cmd/api

test:
	go test ./... -race

lint:
	golangci-lint run

tidy:
	go mod tidy

migrate-up:
	$(MIGRATE) -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" up

migrate-down:
	$(MIGRATE) -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" down 1

migrate-new:
	@test -n "$(name)" || (echo "usage: make migrate-new name=create_xxx" && exit 1)
	$(MIGRATE) create -ext sql -dir $(MIGRATIONS_DIR) -seq -digits 6 $(name)

seed:
	go run ./cmd/seed
