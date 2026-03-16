# Simple Makefile to make working with the transfersystem project convenient.
#
# Common targets:
#   make deps        - download Go dependencies
#   make test        - run Go unit tests
#   make run         - run the API server
#   make scenarios   - run the end-to-end scenario script

.PHONY: all deps test run scenarios

all: test

deps:
	cd transfersystem && go mod tidy

test:
	cd transfersystem && go test ./...

run:
	cd transfersystem && go run ./cmd/main.go
