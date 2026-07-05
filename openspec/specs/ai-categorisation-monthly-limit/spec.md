## ADDED Requirements

### Requirement: Monthly Limit Constant

The system SHALL define a package-level constant `categorisationMonthlyLimit = 50` in `internal/usecase/categorisation_usecase.go`. This constant represents the maximum number of AI categorisation runs a user may accumulate in a single calendar month (UTC).

#### Scenario: Constant is defined

- **WHEN** the Go package is compiled
- **THEN** `categorisationMonthlyLimit` is defined as `50` in the categorisation usecase file without compilation errors

---

### Requirement: CountByUserAndMonth Repository Method

The `categorisationLogRepository` interface SHALL include a `CountByUserAndMonth(ctx context.Context, userID string, year, month int) (int, error)` method. The SQL implementation SHALL count rows in `ai_categorisation_logs` where `user_id` matches and `created_at` falls within the given UTC calendar month.

#### Scenario: Count returns zero for user with no logs this month

- **WHEN** `CountByUserAndMonth` is called for a user who has no log entries in the given month
- **THEN** the method returns `0, nil`

#### Scenario: Count returns correct value for user with logs

- **WHEN** `CountByUserAndMonth` is called for a user who has N log entries in the given month
- **THEN** the method returns `N, nil`

#### Scenario: Repository error is propagated

- **WHEN** the database returns an error during the count query
- **THEN** the method returns a non-nil error

---

### Requirement: CountThisMonth Usecase Method

`CategorisationUsecase` SHALL expose a `CountThisMonth(ctx context.Context, userID string) (int, error)` method. It SHALL delegate to `CountByUserAndMonth` using the current UTC year and month from `time.Now().UTC()`.

This method SHALL also be exposed via a minimal `CategorisationCountUsecase` interface (defined in `internal/handler/me.go`) so that `MeHandler` can depend on it without importing the full usecase type.

#### Scenario: CountThisMonth returns the current month's count

- **WHEN** `CountThisMonth` is called for a user with categorisation logs in the current UTC month
- **THEN** the method returns the correct count for that month

#### Scenario: CountThisMonth propagates repository error

- **WHEN** the underlying repository returns an error
- **THEN** `CountThisMonth` returns a non-nil error

---

### Requirement: CategoriseImage Limit Enforcement

`CategorisationUsecase.CategoriseImage` SHALL check the user's monthly categorisation count before calling the agent. If the count is greater than or equal to `categorisationMonthlyLimit`, the method SHALL return `nil` immediately without calling the agent and without creating a log entry. The River worker receives `nil` and considers the job complete, preventing retries.

#### Scenario: CategoriseImage skips agent when limit is reached

- **WHEN** `CategoriseImage` is called and the user's count for the current month is >= 50
- **THEN** the agent service is not called
- **AND** no log entry is created
- **AND** the method returns `nil`

#### Scenario: CategoriseImage proceeds when under the limit

- **WHEN** `CategoriseImage` is called and the user's count for the current month is < 50
- **THEN** the method proceeds to call the agent as normal

#### Scenario: Count query error aborts categorisation

- **WHEN** the count query returns an error before the agent is called
- **THEN** `CategoriseImage` returns the error and the agent is not called
