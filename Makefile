.PHONY: build install clean

build:
	go build -o bonsai .

install:
	go install .

clean:
	rm -f bonsai
