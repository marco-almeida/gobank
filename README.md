# gobank

Bank API implemented with Go's http std lib and PostgreSQL.

It is a REST API that allows the creation of users, accounts related to users, deposits, withdrawals, and transfers.

The project is a work in progress and is being developed as a learning experience and to serve as a reference for future projects.

## Features

- [X] User creation
- [x] Account creation
- [x] Deposits
- [x] Withdrawals
- [ ] Transfers

Technical features:

- [x] Project layout
- [x] Dependency Injection
- [x] Authentication (via middleware)
- [x] Authorization (via middleware)
- [x] Logging (via middleware)
- [x] Persistent storage (with PostgreSQL)
- [x] Secure configuration
- [x] OpenAPI 3 documentation
- [x] Versioning
- [x] Pagination
- [x] Per user rate limiting (via middleware)
- [x] Dockerization (with multi-stage builds)
- [x] Graceful shutdown
- [ ] Database migrations
- [ ] Event streaming along with WebSockets or Server-Sent Events to notify clients of requested actions
- [ ] Caching with Redis/Memcached
- [ ] Testing (with coverage) triggered by CI/CD

## Project Layout

This is an opinionated folder structure for Go projects where scalability and maintainability are the main concerns.

The [Standard Go Project Layout](https://github.com/golang-standards/project-layout/tree/master) was taken into account as well as opinions from the golang community such as [How To Structure Your Golang (API) Projects!?](https://www.youtube.com/watch?v=EqniGcAijDI).

- `cmd`: Entrypoint for this project, where the whole application is configured and executed.
- `build`: Packaging and Continuous Integration.
  - `ci` should contain configurations and scripts for CI.
  - `package` should contain cloud, container (Docker) and OS configurations as well as scripts for packaging.
- `internal`: Private application and library code. This is the code you don't want others importing in their applications or libraries. Note that this layout pattern is enforced by the Go compiler itself. You can't import anything under `internal` from outside the repository.
  - `rest`: API code containing the handlers, middlewares, and routes.
  - `storage`: Database interface and interaction code.
  - `types`: Custom types and related code.
- `pkg`: Library code that's ok to use by external applications.
- `configs`: Configuration file templates or default configs.
- `deploy`: IaaS, PaaS, system and container orchestration deployment configurations and templates.

## Getting Started

### Prerequisites

- [Docker](https://docs.docker.com/get-docker/)
- [Docker Compose](https://docs.docker.com/compose/install/)

#### Run locally

If you want to run the API locally, you will need the following:

- [Go 1.22.x](https://golang.org/dl/)
- [Make](https://www.gnu.org/software/make/)

### Steps

1. Clone the repository.

```sh
git clone https://github.com/marco-almeida/gobank.git
```

2. Set the environment variables in the `configs/.env` file according to the template in `configs/example.env`.

3. Run the containers.

```sh
docker compose -f ./deploy/docker-compose.yml --env-file ./configs/.env up # --build if needed for a new image, -d for detached mode
```

If running the API locally, execute the following command:

```sh
make run
```

Access the API at `http://localhost:3000`.

## Documentation

OpenAPI 3 documentation is available at `https://github.com/marco-almeida/gobank/blob/main/api/openapi.yaml`.
