.PHONY: build test lint check cover bench mutate install clean help

build:
	go build -o wen ./cmd/wen

test:
	go test -race -count=1 ./...

lint:
	golangci-lint run

check: test lint

cover:
	go test -race -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out
	@echo "---"
	@echo "HTML report: go tool cover -html=coverage.out"

bench:
	go test -run=^$$ -bench=. -benchmem -count=6 ./... | tee bench.txt
	@echo "---"
	@echo "Compare with: benchstat old.txt bench.txt"

mutate:
	@command -v gremlins >/dev/null 2>&1 || { echo "Install: go install github.com/go-gremlins/gremlins/cmd/gremlins@latest"; exit 1; }
	gremlins unleash --tags "" ./...

install:
	go install ./cmd/wen

clean:
	rm -f wen coverage.out bench.txt

help:
	@echo "Available targets:"
	@echo "  build    Build the wen binary"
	@echo "  test     Run tests with race detector"
	@echo "  lint     Run golangci-lint"
	@echo "  check    Run tests and lint"
	@echo "  cover    Run tests with coverage report"
	@echo "  bench    Run benchmarks (benchstat-compatible output)"
	@echo "  mutate   Run mutation testing with gremlins"
	@echo "  install  Install wen to GOPATH/bin"
	@echo "  clean    Remove built binary and reports"
	@echo "  help     Show this help"
