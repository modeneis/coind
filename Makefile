#!/usr/bin/env bash

.DEFAULT_GOAL := help
.PHONY: teller test lint check format cover help

PACKAGES_PROVIDER = $(shell find ./providers -type d -not -path '\./packages')
PACKAGES_SERVER = $(shell find ./server -type d -not -path '\./server')

coind: ## Run coind.
	go run cmd/coind/coind.go

test: ## Run tests
	go test ./src/... -timeout=1m -cover

test-race: ## Run tests with -race. Note: expected to fail, but look for "DATA RACE" failures specifically
	go test ./src/... -timeout=2m -race

lint: ## Run linters. Use make install-linters first.
	vendorcheck ./...
	gometalinter --deadline=3m -j 2 --disable-all --tests --vendor \
		-E deadcode \
		-E errcheck \
		-E gas \
		-E goconst \
		-E gofmt \
		-E goimports \
		-E golint \
		-E ineffassign \
		-E interfacer \
		-E maligned \
		-E megacheck \
		-E misspell \
		-E nakedret \
		-E structcheck \
		-E unconvert \
		-E unparam \
		-E varcheck \
		-E vet \
		./...

check: lint ## Run tests and linters

cover: ## Runs tests on ./src/ with HTML code coverage
	@echo "mode: count" > coverage-all.out
	$(foreach pkg,$(PACKAGES),\
		go test -coverprofile=coverage.out $(pkg);\
		tail -n +2 coverage.out >> coverage-all.out;)
	go tool cover -html=coverage-all.out

install-linters: ## Install linters
	go get -u github.com/FiloSottile/vendorcheck
	go get -u github.com/alecthomas/gometalinter
	gometalinter --vendored-linters --install

format:  # Formats the code. Must have goimports installed (use make install-linters).
	# This sorts imports by [stdlib, 3rdpart, mdllife/teller]
	goimports -w -local github.com/modeneis/coind ./src
	# This performs code simplifications
	gofmt -s -w ./src

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

teller-log: ## Run teller. To add arguments, do 'make ARGS="--foo" teller'.
	go build -o coind coind.go ${ARGS}
#	rm -rf teller.log
	nohup ./coind &
	make log;

log:
	tail -f ./nohup.out