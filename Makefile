include .env

DB_URL     = postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable
MIGRATE    = migrate -path db/migrations -database "$(DB_URL)"
GO         = go
BINARY     = financial-planning.exe

.PHONY: dev prod build migrate-up migrate-down migrate-down-all seed seed-fresh

dev:
	air

build:
	$(GO) build -o $(BINARY) .

prod: build
	./$(BINARY)

migrate-up:
	$(MIGRATE) up

migrate-down:
	$(MIGRATE) down 1

migrate-down-all:
	$(MIGRATE) down -all

seed:
	$(GO) run ./db/seeder

seed-fresh:
	$(GO) run ./db/seeder --fresh
