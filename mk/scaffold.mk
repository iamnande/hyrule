.PHONY: new-service
new-service: ## setup: scaffold a new service (prompts for slug, name, description, database)
	@./hack/new-service.sh
