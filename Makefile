SHELL := /usr/bin/env bash

# make: project info (the repo as a whole - see SERVICE_NAME below for the
# one deployable this invocation actually targets)
ORG_NAME         := iamnande
PROJECT_NAME     := hyrule
PROJECT_VERSION  := $(shell cat VERSION)
PROJECT_REPO_URL := github.com/$(ORG_NAME)/$(PROJECT_NAME)
PROJECT_WORKDIR  := $(shell pwd)
PROJECT_COMMIT   := $(shell git rev-parse HEAD | cut -c1-8)
PROJECT_LOG_FMT  := `/bin/date "+%Y-%m-%d %H:%M:%S %z [$(ORG_NAME)-$(PROJECT_NAME)]"`

# make: which service under cmd/ this invocation targets - override per call,
# e.g. `make run SERVICE_NAME=some-other-service`
SERVICE_NAME ?= pings

# make: container engine - prefer docker if present, fall back to podman.
# every target below goes through this rather than hardcoding either.
CONTAINER_ENGINE := $(shell command -v docker >/dev/null 2>&1 && echo docker || echo podman)
COMPOSE           = $(CONTAINER_ENGINE) compose -f stack/compose.yml

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

include mk/bootstrap.mk
include mk/build.mk
include mk/cluster.mk
include mk/database.mk
include mk/run.mk
include mk/scaffold.mk
include mk/test.mk
