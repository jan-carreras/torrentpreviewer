




PHONY: generate
generate:
	cd internal/magnet && go generate

PHONY: test
test: generate
	go test -cover ./...
