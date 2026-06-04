## Purpose

Defines the structural conventions for the Go backend's `internal/` package layout,
covering where cross-cutting infrastructure, HTTP middleware, generic utilities,
and interface definitions must reside.

---

## Requirements

### Requirement: Cross-cutting infrastructure lives in platform/
Config and observability packages SHALL reside under `internal/platform/` and
not at the top level of `internal/`.

#### Scenario: Config is under platform
- **WHEN** the codebase is built
- **THEN** `internal/platform/config` exists and `internal/config` does not

#### Scenario: Observability is under platform
- **WHEN** the codebase is built
- **THEN** `internal/platform/observability` exists and `internal/observability` does not

### Requirement: HTTP middleware lives under handler/
HTTP middleware SHALL reside at `internal/handler/middleware/` and not at the
top level of `internal/`.

#### Scenario: Middleware is under handler
- **WHEN** the codebase is built
- **THEN** `internal/handler/middleware` exists and `internal/middleware` does not

### Requirement: Generic utilities live in pkg/
Packages with no app-specific knowledge SHALL reside under `pkg/` and not
under `internal/`.

#### Scenario: Thumbnail is in pkg
- **WHEN** the codebase is built
- **THEN** `pkg/thumbnail` exists and `internal/thumbnail` does not

### Requirement: Interfaces are defined by their consumer
The usecase layer SHALL define all interfaces it depends on. External service
packages (storage, vision, thumbnail) SHALL contain only implementations.

#### Scenario: StorageService interface is in usecase
- **WHEN** the codebase is built
- **THEN** `StorageService` is declared in `internal/usecase/` and not in `internal/storage/`

#### Scenario: ThumbnailService interface is in usecase
- **WHEN** the codebase is built
- **THEN** `ThumbnailService` is declared in `internal/usecase/` and not in `pkg/thumbnail/`

#### Scenario: VisionService interface is in usecase
- **WHEN** the codebase is built
- **THEN** `VisionService` is declared in `internal/usecase/` and not in `internal/vision/`
