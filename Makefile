.PHONY: build test vet lint fmt check clean

build:
	go build ./...

test:
	go test ./...

vet:
	go vet ./...

lint:
	@command -v golangci-lint >/dev/null 2>&1 || { echo "golangci-lint not installed. Run: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest" >&2; exit 1; }
	golangci-lint run ./...

fmt:
	gofmt -w .
	@if command -v gofumpt >/dev/null 2>&1; then gofumpt -w .; fi

check: fmt lint test

clean:
	rm -f coverage.out
