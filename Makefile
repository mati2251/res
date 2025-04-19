sources :=  $(wildcard pkg/*/*.go) $(wildcard cmd/*.go)

res: $(sources) pkg/db/db.go
	go build -o res ./cmd/...

.PHONY: res
build: res

pkg/db/db.go: db/schema.sql db/queries.sql db/sqlc.yaml
	sqlc -f db/sqlc.yaml generate

.config.yaml:
	cp .config.yaml.example .config.yaml

.PHONY: dev
dev: .config.yaml
	go run ./cmd/... $(ARGS)

postgres_latest.sif:
	apptainer pull --force docker://postgres:latest

.PHONY: dev-db
dev-db: postgres_latest.sif
	apptainer run --fakeroot --writable-tmpfs --env POSTGRES_PASSWORD=postgres postgres_latest.sif

.PHONY: dev-db-schema
dev-db-schema:
	psql postgres postgres -h 127.0.0.1 -p 5432 -f db/schema.sql

.PHONY: dev-db-cli
dev-db-cli:
	pgcli postgres postgres -h 127.0.0.1 -p 5432

.PHONY: format
format:
	go fmt ./...

.PHONY: format-check
format-check:
	@echo "Checking code style with gofmt"
	@! gofmt -l . | read

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: lint
lint:
	golangci-lint run

.PHONY: report
report:
	goreportcard-cli -t 90
