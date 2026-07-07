## ADDED Requirements

### Requirement: Categorisation limit badge on profile avatar

When the authenticated user has `ai_categorisation_enabled = true` and `ai_categorisation_count_this_month >= 50`, `ProfileMenu` SHALL render a small red dot badge overlaid on the avatar. The badge SHALL NOT appear if `ai_categorisation_enabled` is `false` or the count is below 50.

Clicking the `DropdownMenuTrigger` SHALL dismiss the badge for the remainder of the current calendar month. Dismissal SHALL be persisted to `localStorage` under the key `categorisation_limit_dismissed_<YYYY-MM>` (UTC month). On mount, if that key exists for the current UTC month, the badge SHALL NOT be shown regardless of the count.

The `me` query cache (queryKey `['me']`) is the data source; no separate fetch is needed.

#### Scenario: Badge appears when limit is hit and not dismissed

- **WHEN** `ai_categorisation_enabled` is `true` and `ai_categorisation_count_this_month` is `>= 50`
- **AND** the localStorage key for the current month is not set
- **THEN** a red dot badge is visible on the profile avatar

#### Scenario: Badge does not appear when under the limit

- **WHEN** `ai_categorisation_count_this_month` is `< 50`
- **THEN** no badge is rendered on the profile avatar

#### Scenario: Badge does not appear when categorisation is disabled

- **WHEN** `ai_categorisation_enabled` is `false`
- **THEN** no badge is rendered, even if the count is >= 50

#### Scenario: Clicking the trigger dismisses the badge

- **WHEN** the badge is visible and the user clicks the `DropdownMenuTrigger`
- **THEN** the badge is no longer rendered
- **AND** the localStorage key `categorisation_limit_dismissed_<YYYY-MM>` is set for the current UTC month

#### Scenario: Badge stays dismissed within the same month

- **WHEN** the localStorage key for the current UTC month is present
- **THEN** the badge is not rendered even if the count is >= 50 and categorisation is enabled
