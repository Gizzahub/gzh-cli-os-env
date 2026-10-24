.PHONY: build test vet lint fmt fmt-check check clean

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

# Checks what `fmt` would rewrite without writing: gofmt always, gofumpt when
# installed, exactly as `fmt` applies them.
fmt-check:
	@files="$$(gofmt -l .)" || exit 1; \
	if command -v gofumpt >/dev/null 2>&1; then \
		more="$$(gofumpt -l .)" || exit 1; \
		files="$$(printf '%s\n%s\n' "$$files" "$$more" | sed '/^$$/d' | sort -u)"; \
	fi; \
	if [ -n "$$files" ]; then echo "Code is not formatted. Run: make fmt"; printf '%s\n' "$$files"; exit 1; fi

check: fmt lint test

clean:
	rm -f coverage.out
