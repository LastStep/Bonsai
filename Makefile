.PHONY: build install clean test lint fmt tidy

VERSION ?= dev
LDFLAGS := -s -w -X main.version=$(VERSION)

build:
	go build -ldflags "$(LDFLAGS)" -o bonsai .

install:
	go install -ldflags "$(LDFLAGS)" .

clean:
	rm -f bonsai

test:
	go test ./...

lint:
	golangci-lint run ./...

fmt:
	gofmt -s -w .
	goimports -w .

tidy:
	go mod tidy
