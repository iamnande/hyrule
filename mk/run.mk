ENTRYPOINT := ./cmd/main.go

.PHONY: run-signup-api
run-signup-api: ## run: Admin API
	go run -ldflags $(GO_LDFLAGS) $(ENTRYPOINT) -cmd=signup-api
