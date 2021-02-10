




PHONY: generate
generate:
	find ./internal -type d -name "*mocks" -delete
	go generate ./...

PHONY: test
test: generate
	go test -cover ./...
