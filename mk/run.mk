ENTRYPOINT    := ./cmd/main.go
STACK_TIMEOUT := 10

.PHONY: stack-restart
stack-restart: stack-down stack-up ## run: restart the local environment

.PHONY: stack-up
stack-up: ## run: start the local environment
	@echo $(APP_LOG_FMT) "starting local environment"
	@docker compose up --quiet-pull --build --detach --timeout $(STACK_TIMEOUT)

.PHONY: stack-status
stack-status: ## run: show the local environment status
	@echo $(APP_LOG_FMT) "showing local environment status"
	@docker compose ps

.PHONY: stack-down
stack-down: ## run: stop the local environment
	@echo $(APP_LOG_FMT) "stopping local environment"
	@docker compose down --remove-orphans --volumes --timeout $(STACK_TIMEOUT)

.PHONY: run-registration-api
run-registration-api: ## run: Registration API
	go run -ldflags $(GO_LDFLAGS) $(ENTRYPOINT) -cmd=registration-api
