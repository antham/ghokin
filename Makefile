featurePath = $(PWD)
PKGS := $(shell go list ./... | grep -v /vendor)

fmt:
	find . ! -path "./vendor/*" -name "*.go" -exec gofmt -s -w {} \;

gometalinter:
	gometalinter -D gotype -D aligncheck --vendor --deadline=600s --dupl-threshold=200 -e '_string' -j 5 ./...

doc-hunt:
	doc-hunt check -e

run-tests:
	./test.sh

run-quick-tests:
	go test -v $(PKGS)

test-package:
	go test -race -cover -coverprofile=/tmp/ghokin github.com/antham/ghokin/$(pkg)
	go tool cover -html=/tmp/ghokin -o /tmp/ghokin.html
