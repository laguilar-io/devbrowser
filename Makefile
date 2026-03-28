.PHONY: build install test lint clean

build:
	go build -o bin/devbrowser ./cmd/devbrowser

install:
	go install ./cmd/devbrowser

test:
	go test ./...

lint:
	go vet ./...

clean:
	rm -rf bin/
