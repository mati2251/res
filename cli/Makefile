sources :=  $(wildcard pkg/*/*.go) $(wildcard cmd/*.go)

res: $(sources)
	go build -o res ./cmd/...

.PHONY: res
build: res

.PHONY: test
test:
	go test -v ./...

.config.yaml:
	cp .config.yaml.example .config.yaml

.PHONY: dev
dev: .config.yaml
	go run ./cmd/... $(ARGS)

postgres_latest.sif:
	apptainer pull --force docker://postgres:latest

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
