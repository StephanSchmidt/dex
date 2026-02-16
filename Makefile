.PHONY: build test lint check nilcheck sec upgrade-deps install clean

build: go-imports
	go build .

test:
	go test ./...

go-imports:
	go tool goimports -w .

lint:
	go vet ./...
	go tool staticcheck ./...
	go tool golangci-lint run ./...
	go tool nilaway ./...

sec:
	go tool gosec ./...
	go tool govulncheck ./...

check: lint sec test

upgrade-deps:
	go get -u ./...
	go mod tidy
	go test ./...

install: build check
	go install .

clean:
	go clean -cache -i
