# manual-trash-delete

## Purpose

Allows authenticated users to permanently delete individual trashed images or empty their entire trash. Covers the repository method, usecase methods, and HTTP handler endpoints for manual trash deletion.

## Requirements

### Requirement: ListAllTrashed retrieves all trashed images for a user without pagination

The system SHALL provide a `ListAllTrashed(ctx context.Context, userID string) ([]*domain.Image, error)` method on `ImageRepository` that returns all soft-deleted image records for the given user, with no cursor or limit. It SHALL return an empty slice (not an error) when the user has no trashed images.

#### Scenario: Returns all trashed images for the user

- **WHEN** `ListAllTrashed` is called for a user with trashed images
- **THEN** it returns all soft-deleted image records belonging to that user

#### Scenario: Returns empty slice when user has no trashed images

- **WHEN** `ListAllTrashed` is called for a user with no trashed images
- **THEN** it returns an empty slice and no error

---

### Requirement: DeleteFromTrash permanently deletes a single trashed image

The system SHALL provide a `DeleteFromTrash(ctx context.Context, id uuid.UUID, userID string) error` method on `TrashUsecase` (not `ImageUsecase`) that permanently removes a single soft-deleted image owned by the given user.

The method SHALL:
1. Fetch the image via `GetDeletedByID` on `ImageRepository`; return a not-found error if it does not exist in trash
2. Attempt to delete the R2 object at `r2_path` (synchronous, best-effort; log warn on failure, continue)
3. Attempt to delete the R2 object at `thumbnail_path` if not nil (synchronous, best-effort; log warn on failure, continue)
4. Hard-delete the DB record

#### Scenario: Successfully deletes a trashed image

- **WHEN** `DeleteFromTrash` is called with a valid ID belonging to a trashed image owned by the user
- **THEN** the R2 object is deleted, the thumbnail R2 object is deleted (if present), and the DB record is permanently removed

#### Scenario: Returns not-found when image is not in trash

- **WHEN** `DeleteFromTrash` is called with an ID that is not in trash or belongs to another user
- **THEN** it returns a not-found error and performs no deletions

#### Scenario: R2 delete failure does not block hard delete

- **WHEN** `DeleteFromTrash` is called and the R2 object deletion fails
- **THEN** the error is logged at warn level and the DB record is still hard-deleted

---

### Requirement: EmptyTrash permanently deletes all trashed images for a user

The system SHALL provide an `EmptyTrash(ctx context.Context, userID string) error` method on `TrashUsecase` (not `ImageUsecase`) that permanently removes all soft-deleted images owned by the given user.

The method SHALL:
1. Fetch all trashed images via `ListAllTrashed`
2. Hard-delete all DB records (synchronous; determines when 204 is returned to the caller)
3. For each image: enqueue a River job (`R2DeleteArgs`) carrying `r2_path` and `thumbnail_path`
4. Return nil if the user has no trashed images (no-op)

R2 object deletion is asynchronous — it is handled by `R2DeleteWorker` after the DB records are gone. The 204 response is returned after DB commit, not after R2 cleanup.

#### Scenario: Successfully empties all trashed images

- **WHEN** `EmptyTrash` is called for a user with trashed images
- **THEN** all trashed DB records are permanently removed
- **AND** a River job is enqueued per image carrying the R2 path and thumbnail path

#### Scenario: No trashed images results in no-op

- **WHEN** `EmptyTrash` is called for a user with no trashed images
- **THEN** no deletions are performed and no error is returned

#### Scenario: R2 job enqueue proceeds for all images regardless of individual enqueue errors

- **WHEN** `EmptyTrash` is called and enqueueing a River job for one image fails
- **THEN** the error is logged at warn level and the loop continues, attempting to enqueue jobs for remaining images

---

### Requirement: DELETE /images/trash/:id endpoint permanently deletes a single trashed image

The system SHALL expose `DELETE /images/trash/:id` as an authenticated endpoint that calls `DeleteFromTrash` on `TrashUsecase` and returns `204 No Content` on success.

It SHALL return `404 Not Found` if the image is not in the authenticated user's trash.

#### Scenario: Returns 204 on successful deletion

- **WHEN** an authenticated user sends `DELETE /images/trash/:id` with a valid trashed image ID
- **THEN** the response status is `204 No Content` and the image is permanently deleted

#### Scenario: Returns 404 when image is not in trash

- **WHEN** an authenticated user sends `DELETE /images/trash/:id` with an ID not in their trash
- **THEN** the response status is `404 Not Found`

#### Scenario: Returns 400 when ID is not a valid UUID

- **WHEN** an authenticated user sends `DELETE /images/trash/:id` with a malformed ID
- **THEN** the response status is `400 Bad Request`

---

### Requirement: DELETE /images/trash endpoint permanently deletes all trashed images for the user

The system SHALL expose `DELETE /images/trash` as an authenticated endpoint that calls `EmptyTrash` on `TrashUsecase` and returns `204 No Content` on success, including when the user has no trashed images.

#### Scenario: Returns 204 after emptying trash

- **WHEN** an authenticated user sends `DELETE /images/trash` and has trashed images
- **THEN** the response status is `204 No Content` and all trashed images are permanently deleted

#### Scenario: Returns 204 when trash is already empty

- **WHEN** an authenticated user sends `DELETE /images/trash` and has no trashed images
- **THEN** the response status is `204 No Content` and no deletions are performed

---

### Requirement: ProcessR2Delete handles async R2 object deletion

The system SHALL provide a `ProcessR2Delete(ctx context.Context, r2Path string, thumbnailPath *string) error` method on `TrashUsecase` that deletes the R2 object at `r2Path` and, if `thumbnailPath` is non-nil, deletes the object at `thumbnailPath`.

This method is called by `R2DeleteWorker` for jobs enqueued by `EmptyTrash`. Both deletions are attempted; a failure on either is returned as an error (causing River to retry the job).

#### Scenario: Deletes R2 object and thumbnail

- **WHEN** `ProcessR2Delete` is called with a non-nil thumbnail path
- **THEN** both the R2 object and the thumbnail object are deleted from storage

#### Scenario: Deletes R2 object only when no thumbnail

- **WHEN** `ProcessR2Delete` is called with a nil thumbnail path
- **THEN** only the R2 object is deleted

#### Scenario: Returns error on storage failure to trigger River retry

- **WHEN** `ProcessR2Delete` is called and `store.DeleteObject` returns an error
- **THEN** the error is returned so River retries the job

---

### Requirement: R2DeleteArgs carries R2 paths for async deletion

The system SHALL define `R2DeleteArgs` in `internal/usecase/job_args.go`:

```go
type R2DeleteArgs struct {
    R2Path        string  `json:"r2_path"`
    ThumbnailPath *string `json:"thumbnail_path"`
}

func (R2DeleteArgs) Kind() string     { return "r2_delete" }
func (R2DeleteArgs) MaxAttempts() int { return 5 }
```

#### Scenario: R2DeleteArgs satisfies JobArgs interface

- **WHEN** the Go package is compiled
- **THEN** `R2DeleteArgs` implements `usecase.JobArgs` without compilation errors

---

### Requirement: R2DeleteWorker processes async R2 deletion jobs

The system SHALL define `R2DeleteWorker` in `internal/worker/r2_delete.go` that processes `R2DeleteArgs` jobs by calling `trashUsecase.ProcessR2Delete`. It SHALL use River's default retry backoff. Max 5 attempts (as defined by `R2DeleteArgs.MaxAttempts()`).

#### Scenario: Worker calls ProcessR2Delete with job args

- **WHEN** River dispatches an `R2DeleteArgs` job to `R2DeleteWorker`
- **THEN** the worker calls `ProcessR2Delete(ctx, job.Args.R2Path, job.Args.ThumbnailPath)`
