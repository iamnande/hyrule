COVERAGE_DIR := $(APP_WORKDIR)/coverage

# unit coverage
UNIT_DIR      := $(COVERAGE_DIR)/unit
UNIT_WEBPAGE  := $(UNIT_DIR)/index.html
UNIT_COVERAGE := $(UNIT_DIR)/coverage.txt

.PHONY: test-lint
test-lint: ## test: run the linter
	@golangci-lint run -v ./... --max-issues-per-linter=0 --max-same-issues=0

.PHONY: test-unit
test-unit: ## test: run unit test suite
	@echo $(APP_LOG_FMT) "executing unit test suite"
	@mkdir -p $(UNIT_DIR)
	@go test -v \
		-race \
		-count=1 \
		-covermode=atomic \
		-coverpkg=./... \
		-coverprofile=$(UNIT_COVERAGE) \
		./internal/... ./cmd/...
	@go tool cover -func=$(UNIT_COVERAGE)
	@go tool cover -html=$(UNIT_COVERAGE) -o $(UNIT_WEBPAGE)
