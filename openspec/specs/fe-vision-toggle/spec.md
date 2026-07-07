# fe-vision-toggle Specification

## Purpose
Defines the settings UI for toggling AI vision labelling on/off, including how the toggle reflects the authenticated user's `vision_enabled` state and how it persists changes via `PATCH /me`.

## Requirements

### Requirement: AI vision toggle in settings

The settings modal's `AdvancedSection` SHALL render a `Switch` reflecting the authenticated user's `vision_enabled` value, allowing the user to toggle it. Toggling the switch SHALL call `PATCH /me` with the new value. While the request is in flight, the switch SHALL be disabled. On success, the toggle SHALL reflect the persisted value returned by the response without a separate refetch.

#### Scenario: Switch reflects current vision_enabled state

- **WHEN** the settings modal's AI section is rendered
- **THEN** the `Switch` `checked` state matches the `vision_enabled` value from `GET /me`
- **AND** the descriptive text reflects whether AI features are active or disabled

#### Scenario: Toggling the switch enables vision labelling

- **WHEN** the user clicks the switch while `vision_enabled` is `false`
- **THEN** a `PATCH /me` request is made with body `{ "vision_enabled": true }`
- **AND** the switch is disabled until the request completes
- **AND** on success, the switch reflects `checked = true` and the description updates to the active state

#### Scenario: Toggling the switch disables vision labelling

- **WHEN** the user clicks the switch while `vision_enabled` is `true`
- **THEN** a `PATCH /me` request is made with body `{ "vision_enabled": false }`
- **AND** on success, the switch reflects `checked = false` and the description updates to the disabled state

#### Scenario: Failed update leaves the toggle unchanged

- **WHEN** the `PATCH /me` request fails
- **THEN** the switch returns to enabled (no longer pending)
- **AND** the `checked` state remains the last-known value from `GET /me` (no optimistic state to roll back)
- **AND** an error toast ("Failed to update settings") is shown

---

### Requirement: AI auto-categorisation toggle in settings

The settings modal's `AdvancedSection` SHALL render a second toggle row beneath the AI folder suggestions row for AI auto-categorisation. The row SHALL include:

- A `Switch` reflecting the authenticated user's `ai_categorisation_enabled` value from the `me` query cache
- A usage counter displaying `"X / 50 this month"` derived from `ai_categorisation_count_this_month`; when the count is 0 the counter text SHALL be rendered in the muted foreground colour
- Toggling the switch SHALL call `PATCH /me` with `{ "ai_categorisation_enabled": <new value> }`
- While the request is in flight, the switch SHALL be disabled
- On success, the toggle and counter SHALL reflect the values returned in the response without a separate refetch
- The switch SHALL be additionally disabled when `vision_enabled` is `false`; in this state a tooltip hint SHALL indicate that the feature requires Smart Features to be enabled

The `Me` type SHALL be extended with `ai_categorisation_count_this_month: number`. The `UpdateMeParams` type SHALL be extended with `ai_categorisation_enabled?: boolean`.

#### Scenario: Toggle reflects current ai_categorisation_enabled state

- **WHEN** the settings modal's AI section is rendered
- **THEN** the `Switch` `checked` state matches `ai_categorisation_enabled` from the `me` query cache

#### Scenario: Counter reflects current month's usage

- **WHEN** the settings modal's AI section is rendered
- **THEN** the counter shows `"<ai_categorisation_count_this_month> / 50 this month"`

#### Scenario: Toggling the switch enables auto-categorisation

- **WHEN** the user clicks the switch while `ai_categorisation_enabled` is `false` and `vision_enabled` is `true`
- **THEN** a `PATCH /me` request is made with body `{ "ai_categorisation_enabled": true }`
- **AND** the switch is disabled until the request completes
- **AND** on success, the switch reflects `checked = true`

#### Scenario: Toggling the switch disables auto-categorisation

- **WHEN** the user clicks the switch while `ai_categorisation_enabled` is `true` and `vision_enabled` is `true`
- **THEN** a `PATCH /me` request is made with body `{ "ai_categorisation_enabled": false }`
- **AND** on success, the switch reflects `checked = false`

#### Scenario: Failed update leaves the toggle unchanged

- **WHEN** the `PATCH /me` request fails
- **THEN** the switch returns to enabled (no longer pending)
- **AND** the `checked` state remains the last-known value from the `me` query cache
- **AND** an error toast ("Failed to update settings") is shown

#### Scenario: Switch is disabled when vision_enabled is false

- **WHEN** the settings modal's AI section is rendered and `vision_enabled` is `false`
- **THEN** the AI auto-categorisation `Switch` is rendered in a disabled state
- **AND** clicking it does not fire a `PATCH /me` request
