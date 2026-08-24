STACK_TIMEOUT := 10

.PHONY: stack-reset
stack-reset: stack-down stack-up ## run: reset the local environment

.PHONY: stack-up
stack-up: ## run: start the local environment, waiting for postgres to accept connections
	@echo $(PROJECT_LOG_FMT) "starting local environment via $(CONTAINER_ENGINE)"
	@$(COMPOSE) up --quiet-pull --build --detach --timeout $(STACK_TIMEOUT)
	@echo $(PROJECT_LOG_FMT) "waiting for postgres"
	@for i in $$(seq 1 30); do \
		$(CONTAINER_ENGINE) exec hyrule-database pg_isready -U hyrule_owner -d hyrule > /dev/null 2>&1 && exit 0; \
		sleep 1; \
	done; \
	echo "postgres never became ready" && exit 1

.PHONY: stack-status
stack-status: ## run: show the local environment status
	@echo $(PROJECT_LOG_FMT) "showing local environment status"
	@$(COMPOSE) ps

.PHONY: stack-down
stack-down: ## run: stop the local environment
	@echo $(PROJECT_LOG_FMT) "stopping local environment"
	@$(COMPOSE) down --remove-orphans --volumes --timeout $(STACK_TIMEOUT)

.PHONY: run
run: ## run: run the target service locally (SERVICE_NAME=name to override)
	@go run -ldflags $(GO_LDFLAGS) ./cmd/$(SERVICE_NAME)
