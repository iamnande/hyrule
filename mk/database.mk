MIGRATIONS_DIR := migrations
DATABASE_URL   ?= postgres://hyrule_owner:hyrule_owner@localhost:5432/hyrule?sslmode=disable

DB_TIMEOUT := 10

.PHONY: db-up
db-up: ## database: start a standalone postgres container for testing (no cluster needed)
	@echo $(PROJECT_LOG_FMT) "starting hyrule-database via $(CONTAINER_ENGINE)"
	@$(CONTAINER_ENGINE) run \
		--name hyrule-database \
		--detach \
		--rm \
		-e POSTGRES_USER=hyrule_owner \
		-e POSTGRES_PASSWORD=hyrule_owner \
		-e POSTGRES_DB=hyrule \
		-p 5432:5432 \
		-v $(PROJECT_WORKDIR)/local/postgres/init:/docker-entrypoint-initdb.d:ro \
		docker.io/postgres:17.2-alpine
	@echo $(PROJECT_LOG_FMT) "waiting for postgres"
	@for i in $$(seq 1 30); do \
		ready=$$($(CONTAINER_ENGINE) logs hyrule-database 2>&1 | grep -c "database system is ready to accept connections"); \
		[ "$$ready" -ge 2 ] && exit 0; \
		sleep 1; \
	done; \
	echo "postgres never became ready" && exit 1

.PHONY: db-status
db-status: ## database: show the standalone postgres container's status
	@$(CONTAINER_ENGINE) ps --filter name=hyrule-database

.PHONY: db-down
db-down: ## database: stop the standalone postgres container
	@echo $(PROJECT_LOG_FMT) "stopping hyrule-database"
	@$(CONTAINER_ENGINE) stop --time $(DB_TIMEOUT) hyrule-database

.PHONY: db-reset
db-reset: db-down db-up ## database: reset the standalone postgres container

.PHONY: db-migrate-up
db-migrate-up: ## database: apply all pending migrations
	@migrate -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" up

.PHONY: db-migrate-down
db-migrate-down: ## database: roll back one migration
	@migrate -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" down 1

.PHONY: db-migrate-create
db-migrate-create: ## database: scaffold a new migration (NAME=add_foo)
	@migrate create -ext sql -dir $(MIGRATIONS_DIR) -seq $(NAME)
