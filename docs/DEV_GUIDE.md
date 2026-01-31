# Dev Guide

* **Task hooks**: `task hooks` to set git hooks path to `.githooks`; pre-commit runs `task lint` + `task test`.
* **Codegen**: `task config`, `task saas:entg`, `task gql:gql`, `task apidash:gql`. Run after schema/config changes.
* **Docker dev**: `task dev` to start Postgres + app (Air hot reload); optional NATS/ClickHouse commented in `.docker/docker-compose.yml`.
* **Database setup**: If we get db not exists error then we can use terminal to create the database: `PGPASSWORD=postgres psql -h localhost -U postgres -p 5432 -c "CREATE DATABASE appdb;"`
* **DB migrate**: `cd saas && go run ./cmd/dbmigrate` (uses PKL config; docker override at `.docker/app.docker.pkl`).
* **CLI commands**: Always run from `saas` directory: `cd saas && go run ./cmd/cli <command>`
* **GraphQL**: GQL `/app/query`, playground `/app/gql?pkey=1234`; apidash `/sa/query` with `sa_api_key` header, playground `/sa/gql?pkey=1234`.
