## ADDED Requirements

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
