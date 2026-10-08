# Bonsai's make targets. There is no install target on purpose: `go install` puts a `bonsai` in front of the installed
# one on the PATH (CLAUDE.md). `make build` writes ./bonsai (ignored by git); agents build into a scratch folder:
#   make build OUT=~/bonsai-checks/bin/bonsai
.PHONY: build clean test lint fmt tidy

VERSION ?= dev
OUT ?= bonsai
LDFLAGS := -s -w -X main.version=$(VERSION)

build:
	go build -ldflags "$(LDFLAGS)" -o $(OUT) ./cmd/bonsai

clean:
	rm -f $(OUT)

test:
	go test ./...

lint:
	golangci-lint run ./...

fmt:
	gofmt -s -w .
	goimports -w .

tidy:
	go mod tidy
