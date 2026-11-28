include .env

MIGRATIONS_DIR = ./cmd/migrations


.PHONY: migrate-create
migrate-create:
	@migrate create -seq -ext sql -dir $(MIGRATIONS_DIR) $(filter-out $@,$(MAKECMDGOALS))

.PHONY: migrate-up
migrate-up:
	@migrate -path=$(MIGRATIONS_DIR) -database="$(DATABASE_URL)" up

.PHONY: migrate-down
migrate-down:
	@migrate -path=$(MIGRATIONS_DIR) -database="$(DATABASE_URL)" down $(filter-out $@,$(MAKECMDGOALS))
