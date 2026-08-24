MIGRATIONS_DIR := migrations
DATABASE_URL   ?= postgres://hyrule_owner:hyrule_owner@localhost:5432/hyrule?sslmode=disable

.PHONY: db-migrate-up
db-migrate-up: ## database: apply all pending migrations
	@migrate -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" up

.PHONY: db-migrate-down
db-migrate-down: ## database: roll back one migration
	@migrate -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" down 1

.PHONY: db-migrate-create
db-migrate-create: ## database: scaffold a new migration (NAME=add_foo)
	@migrate create -ext sql -dir $(MIGRATIONS_DIR) -seq $(NAME)
