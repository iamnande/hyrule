.PHONY: new-service
new-service: ## setup: scaffold a new service (prompts for slug, name, description, database)
	@./local/new-service.sh
