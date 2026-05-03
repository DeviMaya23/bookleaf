## MODIFIED Requirements

### Requirement: Runnable HTTP Server

The project SHALL be runnable with a single command and respond to HTTP requests immediately after startup.

#### Scenario: Developer starts the server

- **WHEN** `go run ./cmd/server` is executed
- **THEN** the server starts and listens on a configurable port (default `8080`)
- **AND** `GET /health` returns `200 OK` with a JSON body containing component statuses

## ADDED Requirements

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
