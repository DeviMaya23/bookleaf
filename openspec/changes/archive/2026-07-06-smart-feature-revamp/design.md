## Context

The app has two AI feature flags: `vision_enabled` (now branded "Smart Features") and `ai_categorisation_enabled` (gated behind it). The original implementation included a third AI surface — a toast-based folder suggestion flow — which polled the image after upload, read `suggested_folder_name` from the `GET /images/:id` response, and showed a toast prompting the user to accept or ignore. This flow was replaced by smart search and AI auto-categorisation but never cleaned up, leaving dead code across the handler, usecase, frontend hook, and API surface.

The settings UI renders both toggles without any visual relationship between them, even though `ai_categorisation_enabled` only has effect when `vision_enabled` is on (enforced by the backend's `ProcessVisionLabelling` early-return).

## Goals / Non-Goals

**Goals:**
- Remove all code belonging to the folder suggestion toast flow
- Remove `suggested_folder_name` from the API surface (GET /images/:id response)
- Gate the categorisation toggle in the settings UI so it is visually disabled when `vision_enabled` is off
- Update toggle copy and tooltip text to reflect the "Smart Features" framing

**Non-Goals:**
- Changing the backend enforcement of the `vision_enabled` → `ai_categorisation_enabled` dependency (it already works correctly)
- Modifying the Vision API labelling worker or job queue
- Updating `AiNotesPage.tsx` copy (handled manually)
- Adding any new AI surfaces or capabilities

## Decisions

### Decision: Delete `useVisionSuggestion.ts` entirely rather than gut it

The hook's only purpose is polling + toast + accept-suggestion call. There is no reusable logic worth preserving. Deleting the file is cleaner than leaving an empty export.

### Decision: Remove `onUploadSuccess` prop from `UploadModal`

The prop exists solely to pass `checkVision` in from `AppLayout`. With `checkVision` gone, the prop has no remaining callers or purpose. Removing it keeps the component's API minimal.

### Decision: Keep `suggested_folder_name` removal to the `GetImageResponse` type only (not the DB column)

The `ai_labels` column is the source of truth; `suggested_folder_name` was always derived at read time in the usecase. No migration is needed — removing the `suggestedFolderName()` helper and the field from the response struct is sufficient.

### Decision: Disable (not hide) the categorisation toggle when `vision_enabled` is off

Hiding the toggle would obscure the feature's existence. Rendering it disabled with a tooltip hint ("Requires Smart Features to be enabled") lets users discover it without being able to accidentally enable it while vision is off. The toggle's `disabled` prop is already wired through; this just adds a new condition.

### Decision: No server-side enforcement of the toggle gate

The backend already silently no-ops `ai_categorisation_enabled` when `vision_enabled` is false in `ProcessVisionLabelling`. Adding a PATCH /me validation that rejects `ai_categorisation_enabled: true` when `vision_enabled` is false would be extra complexity with no real benefit — the client gate is sufficient.

## Risks / Trade-offs

**[Risk] Tests referencing `suggested_folder_name` in mock return values** → These are compile-time failures if the field is removed from the type. Updating them is mechanical but spread across several test files (`BatchUploadModal.test.tsx`, `upload.test.ts`, `UploadModal.test.tsx`, handler tests). Grep for the field name before marking the task done.

**[Risk] `AcceptSuggestion` is still part of the `UploadUsecase` interface in `handler/image_upload.go`** → Removing it from the interface and the concrete implementation must be done together; the compiler will catch any missed call sites.

**[Trade-off] Categorisation switch disabled state may be surprising** → A user who has `ai_categorisation_enabled: true` in the DB but turns off `vision_enabled` will see the switch visually disabled but still reflecting their stored `true` value. This is the correct UX — the value is preserved, the feature just can't fire. No data is lost.

## Migration Plan

This is a pure removal — no data migrations, no deploy coordination needed. The `suggested_folder_name` field was computed at runtime from `ai_labels`; removing the computation has no effect on stored data.

Rollout order within a single deploy:
1. Backend: remove endpoint and field from response
2. Frontend: remove hook, prop, and field from types; update settings UI

Both can ship together since the frontend already handled `suggested_folder_name` being `null`.
