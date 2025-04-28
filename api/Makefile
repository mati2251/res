include Example.mk

postgres_latest.sif:
	apptainer pull --force docker://postgres:latest

.PHONY: dev-db
dev-db: postgres_latest.sif
	apptainer run --fakeroot --writable-tmpfs --env POSTGRES_PASSWORD=postgres postgres_latest.sif

.PHONY: dev-db-cli
dev-db-cli:
	pgcli res_dev postgres -h 127.0.0.1 -p 5432

.PHONY: db-schema
db-schema:
	mix ecto.create
	mix ecto.migrate

.PHONY: start
start:
	mix phx.server
