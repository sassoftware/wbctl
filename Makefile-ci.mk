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

## release-pr: create a release PR for the latest commit on main
.PHONY: release-pr
release-pr: ROOT    ?= $(shell git rev-parse --show-toplevel)
release-pr: VERSION ?= $(shell $(GOTAGGER))
release-pr: | $(GIT_CHGLOG) $(GOTAGGER)
	@git checkout main
	@git pull
	@git checkout -b "release-$(VERSION)"
	@cd $(ROOT) && $(GIT_CHGLOG) -o CHANGELOG.md --next-tag $(VERSION) --sort=semver
	@git add CHANGELOG.md
	@git commit -s -m "release: $(VERSION)"
	@git push origin "release-$(VERSION)"


COMMIT_TYPE := $(shell git log --pretty=format:'%s' -1 | cut -f1 -d:)
## release-tag: push a release tag if the latest commit is a release commit
.PHONY: release-tag
release-tag: COMMIT_DESC := $(shell git log --pretty=format:'%s' -1 | awk '{print $$2}')
release-tag:
ifeq ($(COMMIT_TYPE),release)
	@git tag "$(COMMIT_DESC)"
	@git push origin "$(COMMIT_DESC)"
else
	@echo Skipping non-release commit
endif
