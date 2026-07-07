## 1. Fix SSE cache invalidation

- [x] 1.1 In `frontend/src/app-shell/useSSEEvents.ts`, add `queryClient.invalidateQueries({ queryKey: ['me'] })` to the `categorisation_complete` branch alongside the existing invalidations

## 2. Tests

- [x] 2.1 Create `frontend/src/app-shell/useSSEEvents.test.ts` — write a test verifying that on a `categorisation_complete` event, the `['me']` query is invalidated (in addition to the existing `['folders']`, `['images']`, `['image', imageId]`, and `['folder']` queries)

## 3. Build & lint

- [x] 3.1 Run `npm run build` from `frontend/` and fix any errors
- [x] 3.2 Run `npm run lint` from `frontend/` and fix any errors
