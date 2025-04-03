sources :=  $(wildcard pkg/*/*.go) $(wildcard cmd/*.go)

res: $(sources) pkg/db/db.go
	go build -o res ./cmd/...

.PHONY: res
build: res

pkg/db/db.go:
	sqlc -f db/sqlc.yaml generate

.config.yaml:
	cp .config.yaml.example .config.yaml

.PHONY: dev
dev: .config.yaml
	go run ./cmd/... $(ARGS)

.PHONY: dev-db
dev-db: .config.yaml
	apptainer instance start docker://postgres postgres

.PHONY: dev-db-stop
dev-db-stop:
	apptainer instance stop postgres

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
