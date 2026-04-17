.PHONY: build install clean

VERSION ?= dev
LDFLAGS := -s -w -X main.version=$(VERSION)

build:
	go build -ldflags "$(LDFLAGS)" -o bonsai ./cmd/bonsai

install:
	go install -ldflags "$(LDFLAGS)" ./cmd/bonsai

clean:
	rm -f bonsai
