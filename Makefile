SHELL := /bin/bash

.PHONY: all build test vet fmt tidy

all: build test

build:
	go build ./...

test:
	go test -v -race ./...

vet:
	go vet ./...

fmt:
	go fmt ./...

tidy:
	go mod tidy
