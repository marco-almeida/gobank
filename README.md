# mybank

Bank API implemented with Golang's Gin, PostgreSQL, and Redis.

It is a REST API that allows for the creation of users, accounts related to users, deposits, withdrawals, and transfers.

This project can be considered as a refactor and extension of techschool's course [Backend Master Class [Golang, Postgres, Redis, Gin, gRPC, Docker, Kubernetes, AWS, CI/CD]
](https://www.udemy.com/course/backend-master-class-golang-postgresql-kubernetes/).

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
- [x] Per-user rate limiting (via middleware)
- [x] Dockerization (with multi-stage builds)
- [x] Graceful shutdown
- [x] Database migrations
- [ ] Event streaming along with WebSockets or Server-Sent Events to notify clients of requested actions
- [ ] Caching with Redis/Memcached
- [ ] Testing (with coverage) triggered by CI/CD

## Project Layout

This is an opinionated folder structure for Go projects where scalability and maintainability are the main concerns.

The [Standard Go Project Layout](https://github.com/golang-standards/project-layout/tree/master) was taken into account as well as opinions from the golang community.

- `cmd`: Entrypoint for this project, where the whole application is configured and executed.
- `build`: Packaging and Continuous Integration.
  - `ci` should contain configurations and scripts for CI. In this case, github actions is used for continuous integration, so this folder is not used.
  - `package` should contain cloud, container (Docker) and OS configurations as well as scripts for packaging.
- `internal`: Domain specific errors and models. Private application and library code. This is the code you don't want others importing in their applications or libraries. Note that this layout pattern is enforced by the Go compiler itself.
  - `handler`: API code containing the handlers.
  - `config`: Configuration code.
  - `pkg`: Code shared by the internal packages.
  - `middleware`: API code containing middlewares.
  - `postgresql`: PostgreSQL interaction code.
  - `service`: Business logic code called by the handlers.
- `api`: OpenAPI/Swagger specs, JSON schema files, protocol definition files.

## Architecture

Layered/Onion Architecture. Before server creation, the layers are instantiated and configured using dependency injection.

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
git clone https://github.com/marco-almeida/mybank.git
```

2. Create a `development.env` file in the project's root directory according to the template in `example.env`.

*If the environment value MYBANK_ENV is set, the file with name ${MYBANK_ENV}.env will be used instead of `development.env`.*

3. Run the containers.

```sh
docker compose --env-file ./development.env up # --build if needed for a new image, -d for detached mode
```

If running the API locally, execute the following command:

```sh
make run
```

Access the API at <http://localhost:3000>.

## Documentation

OpenAPI 3 documentation is available at <https://github.com/marco-almeida/mybank/blob/main/api/openapi.yaml>.
