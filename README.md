# bookleaf

## CI

This repository runs backend Go tests on pull requests targeting `main` via GitHub Actions (`.github/workflows/pr-go-tests.yml`).

- Trigger: `pull_request` to `main`
- Test command: `go test -v ./...` from `backend/`
- Integration tests: testcontainers-based tests run in CI using the runner Docker runtime
