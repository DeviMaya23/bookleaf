## 1. API Layer

- [x] 1.1 Add `backfillVisionLabels(getToken)` to `features/auth/lib/me.ts` — calls `POST /me/vision/backfill`, returns `{ enqueued: number }`, throws on non-OK

## 2. Confirm Dialog Component

- [x] 2.1 Create `features/settings/components/VisionBackfillConfirmDialog.tsx` with props `open`, `onCancel`, `onConfirm`, `isPending` — follows the Dialog pattern from `DeleteFolderDialog`; Enable button disabled when `isPending`

## 3. AdvancedSection Wiring

- [x] 3.1 Add `confirmOpen` boolean state to `AdvancedSection`; change the vision switch `onCheckedChange` so toggling on sets `confirmOpen = true` instead of calling the mutation directly
- [x] 3.2 Add a `useMutation` for the enable+backfill sequence: on mutate, call `updateMe({ vision_enabled: true })` then `backfillVisionLabels`; on success close modal and update query cache; on `PATCH /me` failure show error toast and close modal; on backfill-only failure show warning toast, close modal, and update cache with the enabled state returned from `PATCH /me`
- [x] 3.3 Render `VisionBackfillConfirmDialog` in `AdvancedSection`, wired to `confirmOpen`, `onCancel` (closes modal), `onConfirm` (fires the sequence mutation), and `isPending`

## 4. Unit Tests

- [x] 4.1 Update `AdvancedSection.test.tsx`: clicking the vision switch when off opens the confirm modal and does not immediately call `PATCH /me`
- [x] 4.2 Cancelling the confirm modal closes it and leaves `vision_enabled` false with no API call
- [x] 4.3 Confirming calls `PATCH /me` then `POST /me/vision/backfill` in sequence; on full success modal closes and switch reflects enabled
- [x] 4.4 When `PATCH /me` fails, error toast is shown, modal closes, switch remains off
- [x] 4.5 When `PATCH /me` succeeds but backfill fails, warning toast is shown, modal closes, switch reflects enabled
- [x] 4.6 Clicking the vision switch when on calls `PATCH /me { vision_enabled: false }` directly with no modal

## 5. Quality

- [x] 5.1 Run `npm run build` from the frontend directory and fix any errors
- [x] 5.2 Run `npm run lint` from the frontend directory and fix any findings
