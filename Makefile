MIGRATE_PATH=./schema
PORT=8080
EXTERNAL_API_URL=externalAPI
LOGGER_LEVEL=debug

DB_CONTAINER_NAME=todo-db
POSTGRES_IMAGE=postgres

DB_USER=postgres
POSTGRES_PASSWORD=password
DB_PORT=5432
DB_NAME=postgres
DATABASE_URL=postgres://$(DB_USER):$(POSTGRES_PASSWORD)@localhost:$(DB_PORT)/$(DB_NAME)?sslmode=disable

.PHONY: all all_with_data create_env run migrate migrate_with_data migrate_down build start_db clean check_db

all: create_env start_db check_db migrate build run

all_with_data: create_env start_db check_db migrate_with_data build run

create_env:
	@echo "DATABASE_URL=$(DATABASE_URL)" > .env
	@echo "EXTERNAL_API_URL=$(EXTERNAL_API_URL)" >> .env
	@echo "PORT=$(PORT)" >> .env
	@echo "LOGGER_LEVEL=$(LOGGER_LEVEL)" >> .env
	@echo "File .env was created"

start_db:
	@echo "Up container with PostgreSQL"
	docker run --name=$(DB_CONTAINER_NAME) \
		-e POSTGRES_PASSWORD=$(POSTGRES_PASSWORD) \
		-e POSTGRES_DB=$(DB_NAME) \
		-p $(DB_PORT):5432 \
		-d --rm $(POSTGRES_IMAGE)

check_db:
	gtimeout 90s bash -c "until docker exec $(DB_CONTAINER_NAME) pg_isready ; do sleep 5 ; done"

migrate:
	@echo "Run migrations"
	migrate -path $(MIGRATE_PATH) -database "$(DATABASE_URL)" up 1

migrate_with_data:
	@echo "Run migrations"
	migrate -path $(MIGRATE_PATH) -database "$(DATABASE_URL)" up

migrate_down:
	@echo "Run migrations"
	migrate -path $(MIGRATE_PATH) -database "$(DATABASE_URL)" down

build:
	@echo "Build go server"
	swag init
	go build main.go

run:
	@echo "Run server"
	./main

clean:
	rm -f main