#!/usr/bin/make -f
FILES		?= $(shell find . -type f -name '*.go' -not -path "./vendor/*")


.PHONY: default
default: help

.PHONY: help
help:   ## show this help
	@echo 'usage: make [target] ...'
	@echo ''
	@echo 'targets:'
	@egrep '^(.+)\:\ .*##\ (.+)' ${MAKEFILE_LIST} | sed 's/:.*##/#/' | column -t -c 2 -s '#'

.PHONY: clean
clean:
	go mod tidy
	go clean

tools:  ## fetch and install all required tools
	go get -u golang.org/x/tools/cmd/goimports


.PHONY: generate
generate: ## generate mocks
	find ./internal -type d -name "*mocks" -exec rm -rf {} +
	go generate ./...

.PHONY: test
test: generate ## run tests
	go test -cover ./... | grep -v "mocks"


.PHONY: fmt
fmt:    ## format the go source files
	go fmt ./...
	goimports -w $(FILES)


.PHONY: mvp
mvp: ## Show pending tasks to be done for MVP
	@grep "\[ \]" TODO | grep mvp

