# server-bootstrap Specification

## Purpose
TBD - created by archiving change go-project-scaffold. Update Purpose after archive.
## Requirements
### Requirement: Runnable HTTP Server

The project SHALL be runnable with a single command and respond to HTTP requests immediately after startup.

#### Scenario: Developer starts the server

- **WHEN** `go run ./cmd/server` is executed
- **THEN** the server starts and listens on a configurable port (default `8080`)
- **AND** `GET /health` returns `200 OK` with a JSON body containing component statuses

### Requirement: Clean Architecture Directory Structure

The project SHALL have a clean architecture directory layout present in version control so contributors can immediately locate where each layer of code belongs.

#### Scenario: Clean architecture directories exist

- **WHEN** the repository is cloned
- **THEN** `internal/domain/`, `internal/usecase/`, `internal/repository/`, and `internal/handler/` directories are present
- **AND** each directory contains a `.gitkeep` file so it is tracked by git without requiring committed code

### Requirement: Kinde Environment Variables

The server SHALL load all required configuration via `config.Load()` at startup before any other initialisation. Required env vars are `KINDE_ISSUER_URL`, `KINDE_AUDIENCE`, `DATABASE_HOST`, `DATABASE_NAME`, `DATABASE_PORT`, `DATABASE_USER`, `DATABASE_PASSWORD`, and `DATABASE_SSLMODE`. If any are missing, the server SHALL fail to start with a clear error message naming the missing variable.

#### Scenario: Server starts with all required env vars present

- **WHEN** `KINDE_ISSUER_URL`, `KINDE_AUDIENCE`, `DATABASE_HOST`, `DATABASE_NAME`, `DATABASE_PORT`, `DATABASE_USER`, `DATABASE_PASSWORD`, and `DATABASE_SSLMODE` are set in the environment
- **THEN** the server starts successfully

#### Scenario: Server fails without required env vars

- **WHEN** any of `KINDE_ISSUER_URL`, `KINDE_AUDIENCE`, `DATABASE_HOST`, `DATABASE_NAME`, `DATABASE_PORT`, `DATABASE_USER`, `DATABASE_PASSWORD`, or `DATABASE_SSLMODE` is missing from the environment
- **THEN** the server exits with a non-zero status code and logs which variable is missing

### Requirement: Protected Route Group

The server SHALL define a protected Echo route group with the Kinde auth middleware applied. All routes requiring authentication SHALL be registered on this group.

The `/health` endpoint SHALL remain outside the protected group.

#### Scenario: Health endpoint is accessible without auth

- **WHEN** `GET /health` is called without an Authorization header
- **THEN** the response is `200 OK`

#### Scenario: Protected routes require auth

- **WHEN** a request is made to any route in the protected group without a valid Bearer token
- **THEN** the response is `401 Unauthorized`

### Requirement: Structured Health Check Response

`GET /health` SHALL return `200 OK` in all cases with a JSON body. The handler SHALL probe DB connectivity (via `SELECT 1`) and R2 connectivity (via a lightweight SDK call) with a 3-second deadline. Each component reports `"ok"` or an error string:

```json
{
  "status": "ok",
  "db": "ok",
  "r2": "ok"
}
```

`status` is `"ok"` only when all components are healthy; otherwise it is `"degraded"`. Individual component values are `"ok"` on success or a short error description on failure. The endpoint SHALL always return HTTP `200` regardless of component health so load balancers do not remove a pod due to a transient dependency failure.

#### Scenario: All components healthy

- **WHEN** `GET /health` is called and both DB and R2 are reachable
- **THEN** the response is `200 OK`
- **AND** the body is `{"status":"ok","db":"ok","r2":"ok"}`

#### Scenario: Database unreachable

- **WHEN** `GET /health` is called and the DB probe fails
- **THEN** the response is still `200 OK`
- **AND** `status` is `"degraded"` and `db` contains a non-empty error description

#### Scenario: R2 unreachable

- **WHEN** `GET /health` is called and the R2 probe fails
- **THEN** the response is still `200 OK`
- **AND** `status` is `"degraded"` and `r2` contains a non-empty error description

#### Scenario: Health endpoint accessible without auth

- **WHEN** `GET /health` is called without an Authorization header
- **THEN** the response is `200 OK`
- **AND** the body contains component statuses
