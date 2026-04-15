.PHONY: build install clean

VERSION ?= dev
LDFLAGS := -s -w -X main.version=$(VERSION)

build:
	go build -ldflags "$(LDFLAGS)" -o bonsai .

install:
	go install -ldflags "$(LDFLAGS)" .

clean:
	rm -f bonsai
