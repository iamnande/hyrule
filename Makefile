SHELL := /usr/bin/env bash

ifneq (,$(wildcard ./.env))
    include .env
    export
endif

# make: app info
ORG_NAME     := iamnande
APP_NAME     := hyrule
APP_VERSION  := $(shell cat VERSION)
APP_REPO_URL := github.com/$(ORG_NAME)/$(APP_NAME)
APP_WORKDIR  := $(shell pwd)
APP_COMMIT   := $(shell git rev-parse HEAD | cut -c1-8)
APP_LOG_FMT  := `/bin/date "+%Y-%m-%d %H:%M:%S %z [$(ORG_NAME)-$(APP_NAME)]"`

.DEFAULT_GOAL := help
.PHONY: help
help: ## display this help screen
	@# top-level targets
	@grep -h -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort -k1 | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
	@# hierarchical remaining targets
	@grep -h -E '^[a-zA-Z0-9_-]+/[a-zA-Z0-9/_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk '{print $$1}' | \
		awk -F/ '{print $$1}' | \
		sort -u | \
	while read section ; do \
		echo; \
		grep -h -E "^$$section/[^:]+:.*?## .*$$" $(MAKEFILE_LIST) | sort -k1 | \
			awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}' ; \
	done

include mk/build.mk
include mk/run.mk
include mk/test.mk
