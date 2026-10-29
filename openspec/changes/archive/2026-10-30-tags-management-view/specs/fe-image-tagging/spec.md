# Spec Delta

## ADDED Requirements

### Requirement: Tags API lib exposes update and delete functions

The system SHALL add two functions to `frontend/src/lib/tags.ts`:
- `updateTag(getToken, id, name)` — calls `PUT /tags/:id` with `{ name }` and returns the updated `Tag`. SHALL throw on any non-2xx response except 409. On a 409 response it SHALL return `null` to signal a duplicate name.
- `deleteTag(getToken, id)` — calls `DELETE /tags/:id`. SHALL throw on any non-2xx response.

#### Scenario: updateTag calls PATCH /tags/:id

- **WHEN** `updateTag` is called with a valid id and name
- **THEN** a PUT request is sent to `/tags/:id` with `{ name }` in the body
- **AND** the updated tag object is returned

#### Scenario: updateTag returns null on 409

- **WHEN** `updateTag` receives a 409 response
- **THEN** `null` is returned rather than throwing

#### Scenario: deleteTag calls DELETE /tags/:id

- **WHEN** `deleteTag` is called with a valid id
- **THEN** a DELETE request is sent to `/tags/:id`
- **AND** the function resolves without error on 204
