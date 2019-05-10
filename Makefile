featurePath = $(PWD)
PKGS := $(shell go list ./... | grep -v /vendor)

fmt:
	find . ! -path "./vendor/*" -name "*.go" -exec gofmt -s -w {} \;

lint:
	golangci-lint run

doc-hunt:
	doc-hunt check -e

run-tests:
	./test.sh

run-quick-tests:
	go test -v $(PKGS)

test-package:
	go test -race -cover -coverprofile=/tmp/ghokin github.com/antham/ghokin/$(pkg)
	go tool cover -html=/tmp/ghokin -o /tmp/ghokin.html

test-all: lint run-tests doc-hunt