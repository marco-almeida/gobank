# mybank

![test workflow](https://github.com/marco-almeida/mybank/actions/workflows/test.yml/badge.svg)
[![codecov](https://codecov.io/gh/marco-almeida/mybank/branch/main/graph/badge.svg)](https://codecov.io/gh/marco-almeida/mybank)
![Go Version](https://img.shields.io/badge/go%20version-%3E=1.22-61CFDD.svg?style=flat-square)

Bank API implemented with Golang's Gin, PostgreSQL, and Redis.

It is a REST API that allows for the creation of users, accounts related to users, deposits, withdrawals, and transfers.

This project can be considered as a refactor and extension of techschool's course [Backend Master Class [Golang, Postgres, Redis, Gin, gRPC, Docker, Kubernetes, AWS, CI/CD]
](https://www.udemy.com/course/backend-master-class-golang-postgresql-kubernetes/).

## Features

- [X] User creation
- [X] Account creation
- [X] Transfers
- [ ] Deposits
- [ ] Withdrawals

Technical features:

- [X] Project layout
- [X] Layered Architecture
- [X] Dependency Injection
- [X] Rest API
- [X] Versioning
- [X] Pagination
- [X] Global Error Handling (via middleware)
- [X] Authentication and Authorization (via middleware)
- [X] Rate Limiting per IP (via middleware)
- [X] Role-based access control
- [X] Persistent storage (with PostgreSQL)
- [X] Secure configuration
- [X] OpenAPI documentation
- [X] Database migrations
- [X] Containerization (using docker multi-stage builds)
- [X] Container Orchestration (using docker compose)
- [X] Graceful shutdown
- [X] Testing (with coverage analysis) triggered by CI/CD

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
  - `middleware`: Middlewares used by the handlers/router.
  - `postgresql`: PostgreSQL interaction code.
  - `service`: Business logic code called by the handlers.
- `api`: OpenAPI/Swagger specs, JSON schema files, protocol definition files.

Before server creation, the layers are instantiated and configured using dependency injection.

Service interfaces are defined in the handler package and implemented in the service package.
Repository interfaces are defined in the service package and implemented in the postgresql package.
This way, the layers are decoupled from each other.

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

OpenAPI documentation is available at <https://github.com/marco-almeida/mybank/blob/main/api/openapi.yaml>.
