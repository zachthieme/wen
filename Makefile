.PHONY: build test lint check install

build:
	go build -o wen ./cmd/wen

test:
	go test -race -count=1 ./...

lint:
	golangci-lint run

check: test lint

install:
	go install ./cmd/wen
