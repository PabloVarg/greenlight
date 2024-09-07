include .env

.PHONY: help
help:
	@ echo "Usage:"
	@ sed -n "s/^##//p" ${MAKEFILE_LIST} | column -t -s ":" | sed -e "s/^/ /"

.PHONY: confirm
confirm:
	@echo -n 'Are you sure? [y/N] ' && read ans && [ $${ans:-N} = y ]

## run/api: run the cmd/api application
.PHONY: run/api
run/api:
	@ go run ./cmd/api/

## docker/up: run project
.PHONY: docker/up
docker/up:
	@ docker compose up --detach
	@ air

## docker/up: destroy docker services
.PHONY: docker/down
docker/down:
	@ docker compose down
	@ killall -KILL air

## db/psql: connect to the database using psql
.PHONY: db/psql
db/psql:
	@ docker compose exec -it db psql -U "${DATABASE_NAME}"

## migrations/up: apply all up database migrations
.PHONY: migrations/up
migrations/up: confirm
	@ echo "Running migrations ..."
	migrate --path ./migrations --database "${DATABASE_DSN}" up

## migrations/create name=$1: create a new database migration
.PHONY: migrations/create
migrations/create:
	@ echo "Creating migration ..."
	migrate create --seq --ext .sql --dir ./migrations "${name}"
