ENTRYPOINT := ./cmd/main.go

.PHONY: run-admin-api
run-admin-api: ## run: Admin API
	go run -ldflags $(GO_LDFLAGS) $(ENTRYPOINT) -cmd=admin-api
