## 1. golangci-lint Configuration

- [x] 1.1 Create `backend/.golangci.yml` with a curated linter set (govet, errcheck, staticcheck, ineffassign, unused)

## 2. Unified CI Workflow

- [x] 2.1 Delete `.github/workflows/pr-go-tests.yml`
- [x] 2.2 Create `.github/workflows/pr-ci.yml` with workflow-level concurrency group (`cancel-in-progress: true`)
- [x] 2.3 Add `detect-changes` job using `dorny/paths-filter@v3` with `backend` and `frontend` path filters
- [x] 2.4 Add `backend-ci` job gated on `needs.detect-changes.outputs.backend == 'true'`: setup Go, run `golangci-lint-action`, run `go test ./...`
- [x] 2.5 Add `frontend-ci` job gated on `needs.detect-changes.outputs.frontend == 'true'`: setup Node with npm cache, run `npm ci`, `npm run lint`, `npm test`, `npm run build`
