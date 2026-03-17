.PHONY: build test lint install

build:
	go build -o wen .

test:
	go test -race -count=1 ./...

lint:
	golangci-lint run

install:
	go install .
