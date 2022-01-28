COMMIT?=$(shell git rev-parse --short HEAD)
BUILD_TIME?=$(shell date -u '+%Y-%m-%d_%H:%M:%S')

export GO111MODULE=on
include configs/app.env
include configs/postgres.env

.DEFAULT_GOAL := run

.PHONY: tools
tools: deps
	go install -tags postgres github.com/golang-migrate/migrate/v4/cmd/migrate@v4.15.1

.PHONY: deps
deps: 
	go mod download

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: test
test: deps
	go test -v -race ./...

.PHONY: build
build:
	go build -ldflags "-X main.buildHash=${COMMIT} -X main.buildTime=${BUILD_TIME}" \
		-o ./bin/billionaire ./cmd/billionaire

.PHONY: run_local
run_local: deps
	set -a && source configs/app.env && go run ./cmd/billionaire

.PHONY: run
run:
	set -a && source configs/app.env && docker-compose -f ./deployments/docker-compose.yml up --build -d

.PHONY: migrate
migrate: tools
	migrate -source file://deployments/migrations -database postgres://${TX_DB_USER}:${TX_DB_PASSWORD}@127.0.0.1:${TX_DB_PORT}/${TX_DB_NAME}?sslmode=disable up
	
	
