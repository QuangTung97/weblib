.PHONY: lint
lint:
	go fmt ./...

.PHONY: test
test:
	go test ./...