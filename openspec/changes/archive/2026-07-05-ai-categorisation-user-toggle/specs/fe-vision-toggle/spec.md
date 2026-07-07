## ADDED Requirements

### Requirement: AI auto-categorisation toggle in settings

The settings modal's `AdvancedSection` SHALL render a second toggle row beneath the AI folder suggestions row for AI auto-categorisation. The row SHALL include:

- A `Switch` reflecting the authenticated user's `ai_categorisation_enabled` value from the `me` query cache
- A usage counter displaying `"X / 50 this month"` derived from `ai_categorisation_count_this_month`; when the count is 0 the counter text SHALL be rendered in the muted foreground colour
- Toggling the switch SHALL call `PATCH /me` with `{ "ai_categorisation_enabled": <new value> }`
- While the request is in flight, the switch SHALL be disabled
- On success, the toggle and counter SHALL reflect the values returned in the response without a separate refetch

The `Me` type SHALL be extended with `ai_categorisation_count_this_month: number`. The `UpdateMeParams` type SHALL be extended with `ai_categorisation_enabled?: boolean`.

#### Scenario: Toggle reflects current ai_categorisation_enabled state

- **WHEN** the settings modal's AI section is rendered
- **THEN** the `Switch` `checked` state matches `ai_categorisation_enabled` from the `me` query cache

#### Scenario: Counter reflects current month's usage

- **WHEN** the settings modal's AI section is rendered
- **THEN** the counter shows `"<ai_categorisation_count_this_month> / 50 this month"`

#### Scenario: Toggling the switch enables auto-categorisation

- **WHEN** the user clicks the switch while `ai_categorisation_enabled` is `false`
- **THEN** a `PATCH /me` request is made with body `{ "ai_categorisation_enabled": true }`
- **AND** the switch is disabled until the request completes
- **AND** on success, the switch reflects `checked = true`

#### Scenario: Toggling the switch disables auto-categorisation

- **WHEN** the user clicks the switch while `ai_categorisation_enabled` is `true`
- **THEN** a `PATCH /me` request is made with body `{ "ai_categorisation_enabled": false }`
- **AND** on success, the switch reflects `checked = false`

#### Scenario: Failed update leaves the toggle unchanged

- **WHEN** the `PATCH /me` request fails
- **THEN** the switch returns to enabled (no longer pending)
- **AND** the `checked` state remains the last-known value from the `me` query cache
- **AND** an error toast ("Failed to update settings") is shown
