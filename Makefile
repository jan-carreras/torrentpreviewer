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
	go test ./...

.PHONY: test-cover
test-cover: generate ## run tests with coverage
	go test -cover ./...

.PHONY: test-fast
test-fast: ## run tests without generating mocks
	go test ./...

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
build: build-clean build-osx

.PHONY: build-clean
build-clean:
	rm -f bin/*

.PHONY: build-linux
build-linux: build-clean bin/linux-torrentprev bin/linux-http-api bin/linux-events

bin/linux-torrentprev:
	CGO_ENABLED=1 GOOS=linux go build --tags "libsqlite3 linux" -o ./bin/linux-torrentprev ./cmd/cli/torrentprev/torrentprev.go

bin/linux-http-api:
	CGO_ENABLED=1 GOOS=linux go build --tags "libsqlite3 linux" -o ./bin/linux-http ./cmd/http/http.go

bin/linux-events:
	CGO_ENABLED=1 go build --tags "libsqlite3 linux" -o ./bin/linux-events cmd/cli/events/events.go

build-osx: bin/darwin-torrentprev bin/darwin-http-api darwin-events

bin/darwin-torrentprev:
	GOOS=darwin go build --tags "libsqlite3 darwin" -o ./bin/darwin-torrentprev ./cmd/cli/torrentprev/torrentprev.go

bin/darwin-http-api:
	GOOS=darwin go build --tags "libsqlite3 darwin" -o ./bin/darwin-http ./cmd/http/http.go

bin/darwin-events:
	CGO_ENABLED=1 go build --tags "libsqlite3 darwin" -o ./bin/darwin-events cmd/cli/events/events.go



