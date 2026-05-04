## Context

The repository currently has no GitHub Actions workflow, so backend tests are not automatically executed on pull requests. The backend includes integration tests that use testcontainers-go and require Docker availability, which means CI must run on a runner that can start containers.

## Goals / Non-Goals

**Goals:**
- Run backend Go tests automatically for pull requests targeting `main`
- Include integration tests in the same CI run (not unit-test-only)
- Ensure runner setup is compatible with testcontainers-go expectations
- Keep workflow simple and deterministic for contributors

**Non-Goals:**
- Adding deployment or release workflows
- Splitting test matrix by OS/Go versions in this change
- Reworking repository test architecture

## Decisions

### Decision 1: Use a single PR workflow scoped to `main`
Create one workflow in `.github/workflows/` triggered by `pull_request` for `main`.

**Rationale:** directly enforces merge gate behavior requested by product/workflow needs.

**Alternative considered:** trigger on all branches; rejected to avoid unnecessary CI usage.

### Decision 2: Run tests from `backend/` using `go test ./...`
Use `working-directory: backend` and execute `go test ./...` so unit and integration tests run together.

**Rationale:** matches current local verification command and exercises full suite.

**Alternative considered:** separate jobs for unit and integration tests; rejected for initial scope simplicity.

### Decision 3: Use GitHub-hosted Ubuntu runner with Docker runtime
Use `ubuntu-latest` and standard `actions/setup-go` for Go toolchain setup.

**Rationale:** GitHub-hosted Ubuntu runners include Docker support required by testcontainers-go.

**Alternative considered:** self-hosted runner; rejected due to operational overhead.

### Decision 4: Add conservative CI hygiene settings
Use dependency caching via setup-go cache and set minimal workflow permissions (`contents: read`).

**Rationale:** keeps runs reasonably fast and security posture tight without complicating setup.

## Risks / Trade-offs

- **Integration tests can be slower/flakier in shared runners** → Mitigation: keep workflow focused, monitor failures, and tune later if needed.
- **Docker/testcontainers behavior may vary over runner updates** → Mitigation: pin workflow steps and add explicit troubleshooting notes if failures appear.
- **Single-job workflow couples all test types** → Mitigation: split jobs in a future change if runtime becomes problematic.

## Migration Plan

1. Add workflow file for PR-to-main test execution.
2. Verify workflow uses backend working directory and `go test ./...`.
3. Ensure CI completes with integration tests that spin test containers.
4. Merge and use required status check in branch protection (repository setting).

Rollback: disable or remove the workflow file if CI blocks development unexpectedly.

## Open Questions

- Should branch protection require this workflow status immediately, or after a short stabilization period?
- Do we want a follow-up workflow for push-to-main to catch post-merge regressions?
