.PHONY: new-service
new-service: ## setup: scaffold a new service (prompts for slug, name, description, cp/dp/both)
	@./hack/new-service.sh
