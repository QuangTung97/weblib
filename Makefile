SHELL := /bin/bash

.PHONY: lint
lint:
	go fmt ./...

.PHONY: test
test:
	go test ./...

.PHONY: run
run:
	source .env && go run examples/googlelogin/main.go
