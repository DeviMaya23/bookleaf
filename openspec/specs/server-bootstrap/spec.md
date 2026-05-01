# server-bootstrap Specification

## Purpose
TBD - created by archiving change go-project-scaffold. Update Purpose after archive.
## Requirements
### Requirement: Runnable HTTP Server

The project SHALL be runnable with a single command and respond to HTTP requests immediately after startup.

#### Scenario: Developer starts the server

- **WHEN** `go run ./cmd/server` is executed
- **THEN** the server starts and listens on a configurable port (default `8080`)
- **AND** `GET /health` returns `200 OK`

### Requirement: Clean Architecture Directory Structure

The project SHALL have a clean architecture directory layout present in version control so contributors can immediately locate where each layer of code belongs.

#### Scenario: Clean architecture directories exist

- **WHEN** the repository is cloned
- **THEN** `internal/domain/`, `internal/usecase/`, `internal/repository/`, and `internal/handler/` directories are present
- **AND** each directory contains a `.gitkeep` file so it is tracked by git without requiring committed code

### Requirement: Kinde Environment Variables

The server SHALL read `KINDE_ISSUER_URL` and `KINDE_AUDIENCE` from the environment at startup. If either is missing, the server SHALL fail to start with a clear error message.

#### Scenario: Server starts with Kinde env vars present

- **WHEN** `KINDE_ISSUER_URL` and `KINDE_AUDIENCE` are set in the environment
- **THEN** the server starts successfully

#### Scenario: Server fails without Kinde env vars

- **WHEN** `KINDE_ISSUER_URL` or `KINDE_AUDIENCE` is missing from the environment
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

