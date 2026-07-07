## MODIFIED Requirements

### Requirement: AI vision toggle in settings

The settings modal's `AdvancedSection` SHALL render a `Switch` reflecting the authenticated user's `vision_enabled` value. When the switch is toggled **off**, the existing direct `PATCH /me` call behaviour is retained. When the switch is toggled **on**, a confirm modal SHALL be shown before any API call is made.

The confirm modal SHALL:
- Explain that image data is sent to Google Vision API for labelling
- State that existing images will be processed in the background and become searchable over time
- Provide a Cancel action and an Enable action
- On Cancel: close the modal, leave `vision_enabled` unchanged, make no API calls
- On Enable: call `PATCH /me { "vision_enabled": true }`, then on success call `POST /me/vision/backfill`

While the Enable action is in flight (both API calls), the Enable button SHALL be disabled. The confirm modal SHALL remain open until both calls complete or fail.

If `PATCH /me` fails, the modal SHALL close, the switch SHALL remain off, and an error toast ("Failed to update settings") SHALL be shown.

If `PATCH /me` succeeds but `POST /me/vision/backfill` fails, vision is already enabled — the switch SHALL reflect `checked = true`, the modal SHALL close, and a warning toast ("Smart Features enabled, but existing images could not be queued for processing. Try again later.") SHALL be shown. Vision is NOT rolled back.

On full success (both calls succeed), the modal SHALL close and the switch SHALL reflect `checked = true` with the description updated to the active state.

#### Scenario: Switch reflects current vision_enabled state

- **WHEN** the settings modal's AI section is rendered
- **THEN** the `Switch` `checked` state matches the `vision_enabled` value from `GET /me`
- **AND** the descriptive text reflects whether AI features are active or disabled

#### Scenario: Toggling the switch on shows the confirm modal

- **WHEN** the user clicks the switch while `vision_enabled` is `false`
- **THEN** the confirm modal is shown
- **AND** no `PATCH /me` call is made yet

#### Scenario: Cancelling the confirm modal leaves vision disabled

- **WHEN** the confirm modal is open and the user clicks Cancel
- **THEN** the modal closes
- **AND** the switch remains `checked = false`
- **AND** no API calls are made

#### Scenario: Confirming the modal enables vision and queues backfill

- **WHEN** the confirm modal is open and the user clicks Enable
- **THEN** a `PATCH /me` request is made with body `{ "vision_enabled": true }`
- **AND** on success, `POST /me/vision/backfill` is called
- **AND** on success of both calls, the modal closes and the switch reflects `checked = true`
- **AND** the description updates to the active state

#### Scenario: Enable button is disabled while in flight

- **WHEN** the user clicks Enable and the requests are in flight
- **THEN** the Enable button is disabled until both calls complete or fail

#### Scenario: PATCH /me failure closes modal with error toast

- **WHEN** the confirm modal is open, the user clicks Enable, and `PATCH /me` fails
- **THEN** the modal closes
- **AND** the switch remains `checked = false`
- **AND** an error toast ("Failed to update settings") is shown

#### Scenario: Backfill failure after successful enable shows warning toast

- **WHEN** `PATCH /me` succeeds but `POST /me/vision/backfill` fails
- **THEN** the modal closes
- **AND** the switch reflects `checked = true` (vision is enabled)
- **AND** a warning toast ("Smart Features enabled, but existing images could not be queued for processing. Try again later.") is shown

#### Scenario: Toggling the switch disables vision labelling

- **WHEN** the user clicks the switch while `vision_enabled` is `true`
- **THEN** a `PATCH /me` request is made with body `{ "vision_enabled": false }` directly (no modal)
- **AND** on success, the switch reflects `checked = false` and the description updates to the disabled state

#### Scenario: Failed disable update leaves the toggle unchanged

- **WHEN** the user clicks the switch to disable and the `PATCH /me` request fails
- **THEN** the switch returns to enabled (no longer pending)
- **AND** the `checked` state remains `true`
- **AND** an error toast ("Failed to update settings") is shown

## ADDED Requirements

### Requirement: backfillVisionLabels API function

The `features/auth/lib/me.ts` module SHALL export a `backfillVisionLabels(getToken: GetToken): Promise<{ enqueued: number }>` function. The function SHALL call `POST /me/vision/backfill` with a valid Bearer token and return the parsed JSON response on success. On a non-OK response it SHALL throw an error.

#### Scenario: Successful backfill call returns enqueued count

- **WHEN** `backfillVisionLabels` is called and the server responds with `202 Accepted` and `{ "enqueued": 5 }`
- **THEN** the function returns `{ enqueued: 5 }`

#### Scenario: Non-OK response throws

- **WHEN** `backfillVisionLabels` is called and the server responds with a non-2xx status
- **THEN** the function throws an error
