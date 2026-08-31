.PHONY: test-lint
test-lint: ## test: run the linter
	@golangci-lint run -v ./... --max-issues-per-linter=0 --max-same-issues=0

.PHONY: test-helm
test-helm: ## test: lint, template, and unittest all helm charts
	@echo $(PROJECT_LOG_FMT) "testing helm charts"
	@for chart in app platform; do \
		helm lint deploy/helm/$$chart; \
		helm template deploy/helm/$$chart > /dev/null; \
	done
	@for values in deploy/values/*/values.yaml; do \
		helm lint deploy/helm/app-platform -f $$values; \
		helm template deploy/helm/app-platform -f $$values > /dev/null; \
	done
	@for chart in app platform app-platform; do \
		helm unittest deploy/helm/$$chart; \
	done

.PHONY: test-unit
UNIT_TEST_COVERAGE_DIR  := $(PROJECT_WORKDIR)/coverage/unit
UNIT_TEST_COVERAGE_PATH := $(UNIT_TEST_COVERAGE_DIR)/coverage.txt
UNIT_TEST_COVERAGE_HTML := $(UNIT_TEST_COVERAGE_DIR)/coverage.html
UNIT_TEST_OPTS :=
ifeq ($(TEST_VERBOSE),true)
	UNIT_TEST_OPTS += -ginkgo.v
endif
test-unit: ## test: execute unit test suite
	@echo $(PROJECT_LOG_FMT) "executing unit test suite"
	@mkdir -p $(UNIT_TEST_COVERAGE_DIR)
	@go test -v \
		-race \
		-count=1 \
		-covermode=atomic \
		-coverpkg=./... \
		-coverprofile=$(UNIT_TEST_COVERAGE_PATH) \
		./go/internal/... $(UNIT_TEST_OPTS)
	@go tool cover -func=$(UNIT_TEST_COVERAGE_PATH)
	@go tool cover -html=$(UNIT_TEST_COVERAGE_PATH) -o $(UNIT_TEST_COVERAGE_HTML)

.PHONY: test-integration
INTEGRATION_TEST_PROCS := 4
INTEGRATION_TEST_OPTS :=
ifeq ($(TEST_VERBOSE),true)
	INTEGRATION_TEST_OPTS += -v
endif
test-integration: ## test: execute integration test suite
	@echo $(PROJECT_LOG_FMT) "executing integration test suite"
	@go tool ginkgo run -r \
		--race \
		--randomize-all \
		--procs=$(INTEGRATION_TEST_PROCS) \
		./go/tests $(INTEGRATION_TEST_OPTS)

.PHONY: test-smoke
SMOKE_BIN := /tmp/hyrule-smoke-$(SERVICE_NAME)
test-smoke: ## test: build+run the target service for real, curl /discovery and /readyz, stop it
	@echo $(PROJECT_LOG_FMT) "smoke testing $(SERVICE_NAME)"
	@go build -ldflags $(GO_LDFLAGS) -o $(SMOKE_BIN) ./go/cmd/$(SERVICE_NAME); \
	$(SMOKE_BIN) & \
	pid=$$!; \
	trap "kill $$pid 2>/dev/null; rm -f $(SMOKE_BIN)" EXIT; \
	up=0; \
	for i in $$(seq 1 20); do \
		curl -sf http://localhost:8000/discovery > /dev/null 2>&1 && { up=1; break; }; \
		sleep 0.25; \
	done; \
	if [ "$$up" != "1" ]; then echo "smoke: service never came up"; exit 1; fi; \
	curl -sf http://localhost:8000/discovery > /dev/null || (echo "smoke: /discovery failed" && exit 1); \
	curl -sf http://localhost:8000/readyz > /dev/null || (echo "smoke: /readyz failed" && exit 1); \
	echo $(PROJECT_LOG_FMT) "smoke test passed"
