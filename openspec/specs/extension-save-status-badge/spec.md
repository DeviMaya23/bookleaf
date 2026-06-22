# extension-save-status-badge

## Purpose

TBD

## Requirements

### Requirement: In-flight save counter

The background service worker SHALL maintain a single in-memory counter of saves currently in flight, incremented at the start of `handleSave` and `handleCapture` and decremented when either function exits, regardless of whether it resolves or throws. The counter SHALL NOT be persisted to storage and SHALL NOT be scoped per tab.

#### Scenario: Counter increments when a save starts

- **WHEN** `handleSave` or `handleCapture` is invoked
- **THEN** the in-flight counter increases by one before any other work in that call begins

#### Scenario: Counter decrements on successful completion

- **WHEN** a save started via `handleSave` or `handleCapture` completes successfully
- **THEN** the in-flight counter decreases by one

#### Scenario: Counter decrements on failure

- **WHEN** a save started via `handleSave` or `handleCapture` fails at any point (auth check, fetch, upload)
- **THEN** the in-flight counter decreases by one, the same as on success

#### Scenario: Concurrent saves are each counted

- **WHEN** a second save starts via any entry point while a first save is still in flight
- **THEN** the in-flight counter reflects both (its value is at least 2) until each individually completes or fails

### Requirement: Toolbar badge reflects in-flight presence only

The extension's toolbar badge SHALL display a static indicator (a dot) whenever the in-flight counter is greater than zero, and SHALL be cleared whenever the counter is zero. The badge SHALL NOT display a count, SHALL NOT change appearance based on save success or failure, and SHALL NOT be affected by which tab originated the save.

#### Scenario: Badge appears when a save starts

- **WHEN** the in-flight counter goes from 0 to 1
- **THEN** the toolbar badge shows the dot indicator

#### Scenario: Badge stays shown while any save remains in flight

- **WHEN** the in-flight counter is greater than 1 (multiple concurrent saves) and one of them completes
- **THEN** the toolbar badge continues showing the dot indicator, unchanged, as long as the counter remains above zero

#### Scenario: Badge clears when the last in-flight save finishes

- **WHEN** the in-flight counter goes from 1 to 0, whether by success or failure of the last remaining save
- **THEN** the toolbar badge is cleared

#### Scenario: Badge does not distinguish save outcome

- **WHEN** a save fails
- **THEN** the badge clearing behaves identically to a successful save clearing it — no distinct error state is shown on the badge
- **AND** the existing error toast (per `extension-save-image`'s Save failure notification requirement) remains the only outcome signal

### Requirement: Badge state is cleared on every service worker cold start

On module load (the same point where context menus are registered via `onInstalled`), the background service worker SHALL explicitly clear the toolbar badge, regardless of any badge state left over from a previous run.

#### Scenario: Stale badge from a killed service worker is cleared on next wake

- **WHEN** the service worker is killed while the in-flight counter was greater than zero (e.g. idle timeout or browser closing mid-save), leaving the badge visibly showing the dot
- **AND** the service worker is later woken for any reason (a new save, the popup opening, browser restart)
- **THEN** the badge is cleared at module load, before reflecting the fresh (zero) in-flight counter

### Requirement: Development build badge is removed

The existing `DEV`-text badge shown in non-production builds SHALL be removed, since it occupies the same badge slot as the in-flight indicator.

#### Scenario: Non-production build no longer shows DEV badge text

- **WHEN** the extension runs in a non-production build
- **THEN** the toolbar badge does not show the text `DEV`
- **AND** the toolbar badge behaves identically to a production build (dot while saves are in flight, cleared otherwise)
