# ci-pr-go-tests Specification

## Purpose
TBD - created by archiving change ci-go-tests-pr-main. Update Purpose after archive.
## Requirements
### Requirement: Pull Request Test Workflow
The repository SHALL include a GitHub Actions workflow that runs backend tests for pull requests targeting the `main` branch.

#### Scenario: Pull request to main triggers workflow
- **WHEN** a pull request is opened, synchronized, or reopened with `main` as the base branch
- **THEN** the workflow starts automatically
- **AND** it executes backend test steps

### Requirement: Backend Go Test Execution
The workflow SHALL run `go test ./...` from the `backend/` directory.

#### Scenario: Test command runs from backend
- **WHEN** the workflow test job executes
- **THEN** it sets `backend/` as the working directory
- **AND** runs `go test ./...`

### Requirement: Integration Test Container Support
The workflow SHALL run on a runner environment that supports Docker so integration tests using testcontainers-go can start required containers.

#### Scenario: Integration tests requiring containers run in CI
- **WHEN** `go test ./...` executes in the workflow
- **THEN** integration tests that start Postgres test containers can run
- **AND** the workflow fails if those tests fail

