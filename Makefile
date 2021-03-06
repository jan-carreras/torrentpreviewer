#!/usr/bin/make -f
FILES		?= $(shell find . -type f -name '*.go' -not -path "./vendor/*")

CGO_ENABLED = 1
ifeq ($(shell uname -s),Linux)
	OSFLAG = linux
endif
ifeq ($(shell uname -s),Darwin)
	OSFLAG = darwin
endif


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

.PHONY: tools
tools:  ## fetch and install all required tools
	go get -u golang.org/x/tools/cmd/goimports
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.37.1

.PHONY: generate
generate: ## generate mocks
	find ./internal -type d -name "*mocks" -exec rm -rf {} +
	go generate ./...

.PHONY: test
test: generate ## run tests
	go test -race ./...

.PHONY: test-cover
test-cover: ## run tests with coverage
	go test -cover -coverprofile=cover.txt -covermode=atomic ./...

.PHONY: test-fast
test-fast: ## run tests without generating mocks
	go test ./...

test-integration: ## run the integration tests without generating mocks
	go test --tags "libsqlite3 ${OSFLAG} integration" ./...

.PHONY: fmt
fmt:    ## format the go source files
	go fmt ./...
	goimports -w $(FILES)

.PHONY: check
check: check-fmt check-no-todos  ## Run linters & gofmt check

.PHONY: check-no-todos
check-no-todos:
	@test -z $(shell echo $(FILES) | xargs grep --color=always TODO | tee /dev/stderr ) || (echo "[ERR] Some Go file includes a TODO comment" && false)

.PHONY: check-fmt
check-fmt:
	@test -z $(shell gofmt -l $(FILES) | tee /dev/stderr) || (echo "[ERR] Fix formatting issues with 'make fmt'" && false)

.PHONY: lint
lint:
	@which golangci-lint > /dev/null 2>/dev/null || (echo "ERROR: golangci-lint not found" && false)
	@golangci-lint run

.PHONY: mvp
mvp: ## Show pending tasks to be done for MVP
	@grep "\[ \]" TODO | grep mvp

.PHONY: build
build: build-clean bin/torrentprev bin/http-api bin/http-events

.PHONY: build-clean
build-clean:
	rm -f bin/*

bin/torrentprev:
	go build --tags "libsqlite3 ${OSFLAG}" -o ./bin/torrentprev ./cmd/cli/torrentprev/torrentprev.go

bin/http-api:
	go build --tags "libsqlite3 ${OSFLAG}" -o ./bin/http ./cmd/http/http.go

bin/http-events:
	go build --tags "libsqlite3 ${OSFLAG}" -o ./bin/events cmd/cli/events/events.go
