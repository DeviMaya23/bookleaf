## ADDED Requirements

### Requirement: CI Compatibility for Repository Integration Tests
SQL repository integration tests SHALL be executable in GitHub Actions runners that provide Docker, without requiring manual local-only setup.

#### Scenario: Repository integration tests run in GitHub Actions
- **WHEN** GitHub Actions runs `go test ./...` in the backend
- **THEN** repository packages using `TestMain` and testcontainers helpers execute successfully on the runner
- **AND** failures in container startup or DB connectivity fail the workflow
