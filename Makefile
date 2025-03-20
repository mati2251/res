sources := $(wildcard pkg/*/*.go) $(wildcard cmd/*.go)

res: $(sources)
	go build -o res ./cmd/...

build: res

format:
	go fmt ./...

format-check:
	@echo "Checking code style with gofmt"
	@! gofmt -l . | read

tidy:
	go mod tidy

lint:
	golangci-lint run

report:
	goreportcard-cli -t 90
