BUILD_DIR := build

GO_VERSION_PACKAGE := $(APP_REPO_URL)/internal/version

GO_LDFLAGS += '
GO_LDFLAGS += -X $(GO_VERSION_PACKAGE).ServiceCommit=$(APP_COMMIT)
GO_LDFLAGS += -X $(GO_VERSION_PACKAGE).ServiceVersion=$(APP_VERSION)
GO_LDFLAGS += -w -s
GO_LDFLAGS += '

.PHONY: build
build: ## build: compile the application
	@go build -ldflags $(GO_LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME) ./cmd/main.go
