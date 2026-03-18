.PHONY: build test lint check install clean help

build:
	go build -o wen ./cmd/wen

test:
	go test -race -count=1 ./...

lint:
	golangci-lint run

check: test lint

install:
	go install ./cmd/wen

clean:
	rm -f wen

help:
	@echo "Available targets:"
	@echo "  build    Build the wen binary"
	@echo "  test     Run tests with race detector"
	@echo "  lint     Run golangci-lint"
	@echo "  check    Run tests and lint"
	@echo "  install  Install wen to GOPATH/bin"
	@echo "  clean    Remove built binary"
	@echo "  help     Show this help"
