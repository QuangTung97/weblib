SHELL := /bin/bash

.PHONY: test
test:
	go fmt ./...
	go vet ./...
	go test ./...

.PHONY: run
run:
	go run examples/googlelogin/main.go
