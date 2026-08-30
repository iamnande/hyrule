.PHONY: run
run: ## run: run the target service locally (SERVICE_NAME=name to override)
	@go run -ldflags $(GO_LDFLAGS) ./go/cmd/$(SERVICE_NAME)
