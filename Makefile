# All tasks are expected be documented with a comment of the form `##
# targetName: usage information`. When in doubt, consult
# https://clarkgrubb.com/makefile-style-guide

MAKEFLAGS += --warn-undefined-variables
SHELL := bash
.SHELLFLAGS := -eu -o pipefail -c
.DEFAULT_GOAL := help
.DELETE_ON_ERROR:
.SUFFIXES:

include .bingo/Variables.mk

.PHONY: help
help:
	@grep '^## ' $(MAKEFILE_LIST) | sed 's/^.*:## //' | sort | column -t -s:

# directories
REPORTDIR = ./build/reports
BUILDDIR = ./build

# commands
GO         ?= go

# flags
COVERFLAGS  = -covermode $(COVERMODE) -coverprofile $(COVEROUT)
COVERMODE   = atomic
COVEROUT    = $(REPORTDIR)/coverage.out
COVERHTML   = $(REPORTDIR)/coverage.html
LINTFLAGS   = --max-issues-per-linter=0 --max-same-issues=0
TESTFLAGS   =

# recipes

## all: lint, test, and build code
.PHONY: all
all: lint test build

## clean: removes generated files and build artifacts
.PHONY: clean
clean:
	$(RM) -r ./build/

## format: format source code
.PHONY: format
format: LINTFLAGS += --fix
format: lint

## build: compile cmd/wbctl to ./build/wbctl
.PHONY: build
build:
	$(GO) build -o $(BUILDDIR)/wbctl ./cmd/wbctl

## usage: build usage documentation
.PHONY: usage
usage: build
	./scripts/gen-usage.sh ./build/wbctl

## copyright: add copyright headers to source files
.PHONY: copyright
copyright:
	./scripts/add-copyright.sh

## licenses: ensure dependency licenses are up to date
.PHONY: licenses
licenses:
	./scripts/update-licenses.sh

## lint: runs linters on source code
.PHONY: lint
lint: | $(GOLANGCI_LINT)
	$(GOLANGCI_LINT) run $(LINTFLAGS)

## test-with-coverage: run tests and generate coverage reports
.PHONY: test-with-coverage
test-with-coverage: TESTFLAGS := $(COVERFLAGS) $(TESTFLAGS)
test-with-coverage: test
	$(GO) tool cover -html $(COVEROUT) -o $(COVERHTML)
	scripts/package-coverage.sh $(COVEROUT) | tee $(REPORTDIR)/package-coverage.txt

## test: run tests
.PHONY: test
test: | $(REPORTDIR)
	$(GO) test $(TESTFLAGS) ./... | tee $(REPORTDIR)/go-test.txt

## version: print the version currently being worked on
.PHONY: version
version: | $(GOTAGGER)
	$(GOTAGGER)

## tools: install support tools
.PHONY: tools
tools:
	bingo get --verbose

$(REPORTDIR):
	mkdir -p $@
