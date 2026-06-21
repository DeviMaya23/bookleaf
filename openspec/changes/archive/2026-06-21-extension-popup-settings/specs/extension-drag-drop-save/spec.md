## MODIFIED Requirements

### Requirement: Drop zone visibility and positioning

The content script SHALL render a transient drop zone, anchored at the pointer's `dragstart` client coordinates, if and only if the drag-to-save setting is enabled (per `getDragEnabled`, default `true`) AND `dragstart` resolved a non-null `srcUrl`. The drag-enabled check SHALL be read fresh from storage on each `dragstart`, not cached at content-script load time, so a setting change takes effect on the very next drag gesture without requiring a page reload. The drop zone SHALL be removed on `dragend`, regardless of whether a `drop` occurred. As a safety net for sites whose own drag-and-drop UI cancels native dragging (e.g. by calling `preventDefault()` on `dragstart`, which prevents `dragend` from ever firing), the content script SHALL also remove the drop zone on `pointerup` and `mouseup`.

#### Scenario: Drop zone appears for a resolvable image drag when drag-to-save is enabled

- **WHEN** drag-to-save is enabled and `dragstart` resolves a non-null `srcUrl`
- **THEN** the content script SHALL render a drop zone near the `dragstart` pointer position

#### Scenario: Drop zone does not appear for an unresolvable drag

- **WHEN** `dragstart` resolves `srcUrl` as `null`
- **THEN** the content script SHALL NOT render a drop zone for that drag gesture

#### Scenario: Drop zone does not appear when drag-to-save is disabled

- **WHEN** drag-to-save is disabled, regardless of whether `dragstart` would otherwise resolve a non-null `srcUrl`
- **THEN** the content script SHALL NOT render a drop zone for that drag gesture

#### Scenario: Drop zone is removed when the drag ends without a drop

- **WHEN** a drag that produced a drop zone ends via `dragend` without a `drop` on the zone (e.g. the user cancels the drag or drops outside the zone)
- **THEN** the drop zone SHALL be removed from the page

#### Scenario: Drop zone is removed via the pointerup/mouseup safety net when dragend never fires

- **WHEN** a drag that produced a drop zone is taken over by the host page's own drag-and-drop library (e.g. dnd-kit, react-dnd) such that native `dragend` never fires
- **THEN** the drop zone SHALL still be removed from the page once `pointerup` or `mouseup` fires
