.PHONY: proto migrate-up migrate-down build run-frontend run-history run-matching mocks test

proto:
	protoc \
		--go_out=./api --go_opt=paths=source_relative \
		--go-grpc_out=./api --go-grpc_opt=paths=source_relative \
		-I api \
		api/workflow.proto

migrate-up:
	go run ./db/cmd/migrate/... up

migrate-down:
	go run ./db/cmd/migrate/... down

build:
	go build ./frontend/... ./history/... ./matching/...

run-frontend:
	go run ./frontend/cmd/...

run-history:
	go run ./history/cmd/...

run-matching:
	go run ./matching/cmd/...

mocks:
	mockery

test:
	go test mini-workflow/frontend/... mini-workflow/history/... mini-workflow/matching/...

.PHONY: up down logs test clean

# Bring up the cluster using docker-compose
up:
	docker compose up -d --build

# Shutdown the cluster
down:
	docker compose down -v

# Tail logs of all services
logs:
	docker compose logs -f

# Run unit tests across the monorepo
test:
	go test ./... -v -count=1

# Clean binaries
clean:
	rm -f bin/*
