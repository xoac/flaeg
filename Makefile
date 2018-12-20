.PHONY: all

GOFILES := $(shell go list -f '{{range $$index, $$element := .GoFiles}}{{$$.Dir}}/{{$$element}}{{"\n"}}{{end}}' ./... | grep -v '/vendor/')
SRCS = $(shell git ls-files '*.go' | grep -v '^vendor/')

default: clean checks test build

test: clean
	go test -v -cover $(PKGS)

dependencies:
	dep ensure -v

clean:
	rm -f cover.out

build:
	go build

checks: check-fmt
	golangci-lint run

check-fmt: SHELL := /bin/bash
check-fmt:
	diff -u <(echo -n) <(gofmt -d $(GOFILES))

fmt:
	gofmt -s -l -w $(SRCS)
