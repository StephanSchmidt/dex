.PHONY: build test lint check nilcheck sec upgrade-deps install clean

build: go-imports
	go build .

test:
	go test ./...

go-imports:
	go run golang.org/x/tools/cmd/goimports -w .

lint:
	go vet ./...
	go run honnef.co/go/tools/cmd/staticcheck ./...
	go run github.com/golangci/golangci-lint/cmd/golangci-lint run ./...

nilcheck:
	go run go.uber.org/nilaway/cmd/nilaway ./...

sec:
	go run github.com/securego/gosec/v2/cmd/gosec ./...
	go run golang.org/x/vuln/cmd/govulncheck ./...

check: lint nilcheck sec test

upgrade-deps:
	go get -u ./...
	go mod tidy
	go test ./...

install:
	go install .

clean:
	go clean -cache -i
