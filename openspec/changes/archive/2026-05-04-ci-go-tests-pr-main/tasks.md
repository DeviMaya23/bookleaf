## 1. GitHub Actions Workflow Setup

- [x] 1.1 Create a workflow file under `.github/workflows/` that triggers on `pull_request` events targeting `main`
- [x] 1.2 Configure a job on `ubuntu-latest` with `actions/checkout` and `actions/setup-go`
- [x] 1.3 Set minimal workflow/job permissions (`contents: read`) and enable Go module caching

## 2. Backend Test Execution in CI

- [x] 2.1 Configure the workflow to run test steps from the `backend/` directory
- [x] 2.2 Run `go test ./...` in CI so both unit and integration tests are exercised
- [x] 2.3 Ensure workflow fails on test failures and surfaces logs for debugging

## 3. Integration Test Container Compatibility

- [x] 3.1 Validate workflow environment supports testcontainers-based integration tests (Docker runtime availability)
- [x] 3.2 Add any required CI env settings for stable testcontainers execution if needed by current tests

## 4. Documentation and Verification

- [x] 4.1 Update repository documentation/comments describing the PR test workflow behavior
- [x] 4.2 Verify workflow syntax and run the backend test command locally before opening PR
