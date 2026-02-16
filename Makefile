.PHONY: build test lint install

build:
	go build .

test:
	go test ./...

lint:
	go vet ./...

install:
	go install .

