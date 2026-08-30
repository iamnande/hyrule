BUILD_DIR := build

GO_VERSION_PACKAGE := $(PROJECT_REPO_URL)/go/internal/lib/version

GO_LDFLAGS += '
GO_LDFLAGS += -X $(GO_VERSION_PACKAGE).ServiceCommit=$(PROJECT_COMMIT)
GO_LDFLAGS += -X $(GO_VERSION_PACKAGE).ServiceVersion=$(PROJECT_VERSION)
GO_LDFLAGS += -X $(GO_VERSION_PACKAGE).RepositoryURL=https://$(PROJECT_REPO_URL)
GO_LDFLAGS += -w -s
GO_LDFLAGS += '

.PHONY: build
build: ## build: compile the target service for the local machine
	@go build -ldflags $(GO_LDFLAGS) -o $(BUILD_DIR)/$(SERVICE_NAME) ./go/cmd/$(SERVICE_NAME)

.PHONY: image-build
image-build: ## build: build the target service's container image
	@echo $(PROJECT_LOG_FMT) "building $(SERVICE_NAME) image via $(CONTAINER_ENGINE)"
	@$(CONTAINER_ENGINE) build \
		--build-arg ORG_NAME=$(ORG_NAME) \
		--build-arg SERVICE_NAME=$(SERVICE_NAME) \
		--build-arg PROJECT_COMMIT=$(PROJECT_COMMIT) \
		--build-arg PROJECT_VERSION=$(PROJECT_VERSION) \
		--build-arg PROJECT_REPO_URL=$(PROJECT_REPO_URL) \
		--build-arg GO_VERSION_PACKAGE=$(GO_VERSION_PACKAGE) \
		--build-arg BUILD_DATETIME=$(shell date -u +%Y-%m-%dT%H:%M:%SZ) \
		-f go/cmd/Dockerfile \
		-t $(ORG_NAME)/$(SERVICE_NAME):$(PROJECT_VERSION) \
		.

.PHONY: image-run
image-run: ## build: run the target service's built image locally
	@$(CONTAINER_ENGINE) run --rm -p 8000:8000 $(ORG_NAME)/$(SERVICE_NAME):$(PROJECT_VERSION)
