STACK_TIMEOUT := 10

.PHONY: stack-reset
stack-reset: stack-down stack-up ## run: reset the local environment

.PHONY: stack-up
stack-up: ## run: start the local environment
	@echo $(PROJECT_LOG_FMT) "starting local environment via $(CONTAINER_ENGINE)"
	@$(COMPOSE) up --quiet-pull --build --detach --timeout $(STACK_TIMEOUT)

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
