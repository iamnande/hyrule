.PHONY: test-lint
test-lint: ## test: run the linter
	@golangci-lint run -v ./... --max-issues-per-linter=0 --max-same-issues=0

.PHONY: test-unit
UNIT_TEST_COVERAGE_DIR  := $(APP_WORKDIR)/coverage/unit
UNIT_TEST_COVERAGE_PATH := $(UNIT_TEST_COVERAGE_DIR)/coverage.txt
UNIT_TEST_COVERAGE_HTML := $(UNIT_TEST_COVERAGE_DIR)/coverage.html
UNIT_TEST_OPTS :=
ifeq ($(TEST_VERBOSE),true)
	UNIT_TEST_OPTS += -ginkgo.v
endif
test-unit: ## test: execute unit test suite
	@echo $(APP_LOG_FMT) "executing unit test suite"
	@mkdir -p $(UNIT_TEST_COVERAGE_DIR)
	@go test -v \
		-race \
		-count=1 \
		-covermode=atomic \
		-coverpkg=./... \
		-coverprofile=$(UNIT_TEST_COVERAGE_PATH) \
		./internal/... $(UNIT_TEST_OPTS)
	@go tool cover -func=$(UNIT_TEST_COVERAGE_PATH)
	@go tool cover -html=$(UNIT_TEST_COVERAGE_PATH) -o $(UNIT_TEST_COVERAGE_HTML)

.PHONY: test-integration
INTEGRATION_TEST_OPTS :=
ifeq ($(TEST_VERBOSE),true)
	INTEGRATION_TEST_OPTS += -ginkgo.v
endif
test-integration: ## test: execute integration test suite
	@echo $(APP_LOG_FMT) "executing integration test suite"
	@go test -v \
		-race \
		-count=1 \
		./tests/... $(INTEGRATION_TEST_OPTS)
