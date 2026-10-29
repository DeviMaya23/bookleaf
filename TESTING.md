# Testing

> Agent-agnostic. Applies to all implementation agents.

---

## Backend

### Philosophy

Test usecase behaviour, not structure. A test that only verifies a method was called, or that an error passed through unchanged, tests implementation — not correctness. Integration tests cover real outcomes; unit tests cover logic that lives exclusively in the usecase.

### Layer coverage

| Layer | Test type |
|---|---|
| `usecase/` | Unit tests |
| `handler/` | Unit tests |
| `repository/` | Integration tests only — no unit tests |

### What to test in usecases

**Happy path** — only if the usecase adds assembly or transformation. If the function delegates to a repo and returns the result unchanged, the happy path is covered by integration tests.

Worth testing:
- Assembling a composite result from multiple sources
- Computing a derived value (path construction, filename, cursor)
- Conditional output based on branching logic

**Exception:** external services not backed by integration tests (e.g. `StorageService`, `VisionService`, `ThumbnailService`). Happy paths that assemble results from these are worth unit testing even without transformation, because there is no integration fallback.

**Error path** — only if the usecase adds behaviour to the error:
- Wrapping: `fmt.Errorf("context: %w", err)`
- Mapping to a domain sentinel: `ErrInvalidFolderName`, `ErrDuplicateTagName`
- Side effects or conditional logic triggered on the error path

Do not write error path tests for pure pass-through (`if err != nil { return nil, err }` with no other logic).

### What not to test

- Pure delegation — function calls one dependency and returns the result unchanged
- Error pass-through — `return nil, err` with no wrapping, mapping, or side effects
- Logging side effects — not verifiable in unit tests

### Test doubles

**Use a fake** when the test requires setting up pre-state in a dependency AND asserting post-state. The signal is: the function reads existing state to decide what to write.

```go
// AcceptSuggestion reads whether a folder exists → branches → writes image folder assignment
// Fake starts with: folder "Nature" exists or doesn't
// Fake ends with:   image is in folder "Nature"
// Assert the outcome, not the call graph
```

**Use a value-return spy** for everything else — single-step operations, external services, and cases where what got called IS the verifiable behaviour (e.g. confirming a no-op skipped a write).

Test doubles implement only the narrow interface defined by the consuming usecase — not the full repository interface. If the usecase defines `type imageCounter interface { CountByFolderID(...) }`, the test double implements only that one method.

### Table-driven tests

Use TDT when all of the following are true:
- 3 or more scenarios
- All cases share identical setup structure
- Cases differ only in inputs, outputs, or expected errors

Use individual `t.Run` blocks when cases need different setup, or when there are fewer than 3 scenarios. Do not stretch TDT to fit cases with varying mock configurations.

### Assertion quality

- Assert the result, not just the error
- Failure scenarios must assert the specific error type or message — not just that an error occurred
- Use `require.ErrorIs` for sentinel errors, `require.ErrorContains` for message matching

### Repository integration tests

Repository tests run against a real database via testcontainers. They verify the **database contract** — what each method promises about its interaction with the schema.

**Always test:**

- **Query correctness** — the right records are returned with the right fields, including preloaded associations where applicable
- **Ownership / user isolation** — a record belonging to user A is not accessible to user B. This is a security property and the most commonly missing test. Methods that scope by `userID` must have a wrong-user case asserting `gorm.ErrRecordNotFound`
- **Soft delete semantics** — soft-deleted records are excluded from standard queries and visible only in explicit trash/deleted queries
- **Constraint behaviour** — unique constraints return the expected error; FK violations are caught
- **Cascade behaviour** — `DeleteWithCascade` and similar operations verify that related rows are actually affected

**Do not test:**

- Infrastructure failures by closing the DB connection — this tests the driver, not the contract. Drop any test that forces an error by calling `sqlDB.Close()`

**Assertion style** — use `require` / `assert` from testify throughout. Do not use bare `if err != nil { t.Fatalf }` checks.

### Test double placement

| Double | Location | Reason |
|---|---|---|
| Fakes (repository interfaces) | `usecase/fakes_test.go` | shared across test files in the package; real logic should not be duplicated |
| Spies (external services) | inline in `*_test.go` | small, vary per scenario, no sharing benefit |
| Spies (handler usecase mocks) | inline in `handler/*_test.go` | per-handler interfaces, nothing to share across files |

### External services

External services (e.g. `StorageService`, `VisionService`) are thin adapters over third-party systems. They delegate directly to an SDK or HTTP client with no domain logic.

- No unit tests on the implementation — correctness depends on the external system, not on code in this repo
- When used as a dependency in usecase tests, always use a value-return spy
- Integration tests only if the adapter contains non-trivial logic worth verifying against the real system

### Handler tests

The handler's job is: parse request → extract auth → call usecase → map errors → serialize response. Tests verify that plumbing is correct.

**Always test:**
- Happy path — assert both the HTTP status code and the response body shape. The handler assembles the HTTP response; that assembly is handler logic even when the usecase result passes through unchanged.
- Each distinct error mapping — `ErrInvalidFolderName → 400`, `gorm.ErrRecordNotFound → 404`, generic error → 500. Each unique mapping is a test case.
- Binding failures — malformed JSON body, invalid UUID path param.

**Request parsing** — handler tests may assert what was passed down to the usecase spy (e.g. `lastDescription`, `lastUpdateParams`) when the interesting logic is in how the handler parses or extracts the request. This is behavioral for the handler layer.

**Test doubles** — always value-return spies for usecase dependencies. Fakes are not used at the handler layer.

**No equivalent of "pure delegation, drop the test"** — even a handler that returns 204 with no body has error mapping worth testing.

---

## Browser Extension

### Philosophy

`extensions/` has no usecase/repository/handler folder split, so the backend's
layer-based test classification doesn't map directly. The underlying criterion
still applies — test logic that lives exclusively in your code; don't test
pass-through; don't unit-test thin adapters over external systems — but the
sorting key is the function's I/O shape, not which folder it's in.

### Classification

| Shape | Example | Unit test? |
|---|---|---|
| Pure logic (no I/O) | `resolveHighResUrl`, `decodeJwtPayload`, `validateResponseShape`, `validateDimension` | Yes |
| Thin browser-adapter (pure passthrough to a browser/fetch API, nothing to assert) | `storage.ts` getters/setters (`getAuth`, `setDarkMode`, etc.) | No |
| Orchestrator (branches, assembles a result, or maps/catches errors around adapters or pure logic) | `resolveImageBlob`, `saveImage`, `handleSave`, `login`/`exchangeCodeForTokens` | Yes, with the I/O boundary mocked |
| Native-browser-API-bound (only logic is wrapping a browser primitive with no test-environment equivalent) | `generateThumbnail` (`OffscreenCanvas`), the `createImageBitmap` call inside `validateCandidate` | No |

When a function mixes a native-browser-API call with branching logic worth
testing (e.g. the original `validateCandidate`), split the branching logic out
into its own pure function and leave a thin, untested orchestrator that owns
only the native call. See `validateResponseShape`/`validateDimension`/
`validateCandidate` in `highResFetch.ts` for the pattern.

### Test doubles

Always use value-return spies (`vi.fn().mockResolvedValue(...)` /
`vi.mocked(x).mockReturnValue(...)`). Fakes are not used in `extensions/` —
nothing here reads pre-existing state from a dependency to decide what to
write the way a backend usecase does against a repository, so there is no
case where a fake's pre-state/post-state setup buys anything over a spy.

### Test setup

- Test runner is vitest, configured in `extensions/vitest.config.ts`
  (`environment: "node"`), kept separate from the build's `vite.config.ts` so
  `vite-plugin-web-extension`'s manifest transform never runs during `npm test`.
- `webextension-polyfill` is mocked per test file via `vi.mock("webextension-polyfill", ...)`
  backed by the shared `createBrowserMock()` helper in `extensions/src/test/browserMock.ts`.
  This is required even for modules that don't call `browser.*` directly in the
  function under test, if importing the module triggers listener registration
  at load time (e.g. `background/index.ts`).
- Test files are colocated next to the unit they cover (`thing.ts` +
  `thing.test.ts`).

### Out of scope (for now)

- `popup/App.tsx` — has orchestration logic (`handleLogin`, `handleLogout`,
  `handleToggleDark`) but no hook extraction, so it isn't unit-testable without
  a structural change. Revisit once there's a decision on applying the
  frontend's hook-extraction-for-testability pattern here.
- `content/index.ts` — DOM-only toast rendering, no branching logic worth
  asserting.
