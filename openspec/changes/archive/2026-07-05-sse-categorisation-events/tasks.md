## 1. BE — SSE Broadcaster

- [x] 1.1 Create `internal/sse/broadcaster.go` with `Event` struct (`Type string`, `Payload json.RawMessage`), `EventBroadcaster` struct (`mu sync.RWMutex`, `clients map[string][]chan Event`), and methods: `Subscribe(userID string) chan Event`, `Unsubscribe(userID string, ch chan Event)`, `Publish(userID string, event Event)` — buffered channels (cap 16), document single-server assumption
- [x] 1.2 Write unit tests for `EventBroadcaster`: publish delivers to all subscriber channels for a user; publish does not deliver to a different user's channels; unsubscribed channel no longer receives events

## 2. BE — GET /events Handler

- [x] 2.1 Create `internal/handler/events.go` with `EventsHandler` struct and `GetEvents` method: subscribe to broadcaster on connect, set SSE headers, loop writing events as `data: <json>\n\n`, unsubscribe on client disconnect (context cancellation)
- [x] 2.2 Write unit test for `EventsHandler.GetEvents`: unauthenticated request returns 401
- [x] 2.3 Create Bruno file `bruno/Get Events SSE.bru` for `GET /events`

## 3. BE — Expose ai_categorisation_enabled on /me

- [x] 3.1 Update `GetMe` handler to include `ai_categorisation_enabled` in the response body
- [x] 3.2 Update `UpdateMe` handler to include `ai_categorisation_enabled` in the response body (read-only — not accepted as a PATCH input)
- [x] 3.3 Update existing `GetMe` handler tests to assert `ai_categorisation_enabled` is present in the response

## 4. BE — Wire Broadcaster into Worker and Routes

- [x] 4.1 Add `broadcaster` field to `CategorisationWorker`; update `NewCategorisationWorker` to accept it; call `broadcaster.Publish` with a `categorisation_complete` event (carrying `image_id`) after `usecase.CategoriseImage` returns `nil`
- [x] 4.2 Write unit test for `CategorisationWorker.Work`: on usecase success, broadcaster receives a publish call with the correct event type and image_id; on usecase error, broadcaster is not called
- [x] 4.3 In `initApp()` in `main.go`, create `EventBroadcaster` unconditionally before workers and handlers; pass it to `NewCategorisationWorker` (inside the `cfg.AnthropicAPIKey != ""` block) and to a new `EventsHandler`; register `GET /events` on the protected route group

## 5. BE — Lint

- [x] 5.1 Run `golangci-lint run ./...` from the backend directory and fix any issues

## 6. FE — Dependencies and Types

- [x] 6.1 Install `@microsoft/fetch-event-source` (`npm install @microsoft/fetch-event-source`)
- [x] 6.2 Add `ai_categorisation_enabled: boolean` to the `Me` interface in `lib/me.ts` (or wherever the Me type is defined)

## 7. FE — useSSEEvents Hook

- [x] 7.1 Create `app-shell/useSSEEvents.ts`: use `fetchEventSource` from `@microsoft/fetch-event-source` to open `GET /events` with `Authorization: Bearer <token>`; on `categorisation_complete` message, invalidate `['folders']`, `['images']`, `['image', imageId]`, and `['folder']`-prefixed queries via `useQueryClient`; open connection only when `me.ai_categorisation_enabled` is true; close connection on unmount

## 8. FE — Mount Hook

- [x] 8.1 Mount `useSSEEvents` in `AppLayout.tsx`

## 9. FE — Build and Lint

- [x] 9.1 Run `npm run build` from the frontend directory and fix any issues
- [x] 9.2 Run `npm run lint` from the frontend directory and fix any issues
