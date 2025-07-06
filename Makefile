SHELL := /bin/bash

.PHONY: lint
lint:
	go fmt ./...
	go vet ./...

.PHONY: test
test:
	go test ./...

.PHONY: run
run:
	go run examples/googlelogin/main.go
