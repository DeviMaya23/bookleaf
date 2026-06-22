## Context

The background service worker (`extensions/src/background/index.ts`) already has a single feedback channel for save outcomes: an in-page toast sent via `browser.tabs.sendMessage` after `persistImage` resolves or rejects (`background/index.ts:330,332`). There is nothing upstream of that — no signal fires when a save starts. For a high-res upload (fetch source image, generate thumbnail, two presigned `PUT`s, a `complete` call) that can take a few seconds, that gap reads as "did anything happen?"

Three independent entry points converge on the same underlying work: `handleDragSaveMessage` → `handleSave`, `handleContextMenuClick` → `handleSave`, and `handleSnipCapturedMessage` → `handleCapture`. Both `handleSave` and `handleCapture` end by calling the shared `persistImage`.

The background script is a Manifest V3 service worker (`manifest.json:31-34`), not a persistent page — it can be killed by the browser's idle timeout or by the browser closing entirely, with no resumable/retry logic anywhere in this chain today. Any in-flight-state design needs to not assume the worker survives to report its own completion.

Several alternatives were considered during exploration and rejected before this design:
- **Numeric count badge** — rejected because a bare number on a toolbar badge needs explaining ("what does '2' mean here?"), and the goal is just alive/not-alive, not precise concurrency.
- **In-page status chip** (a small persistent element rendered into the page via the existing shadow-DOM injection used by the toast/drop-zone) — rejected as feeling like new content appearing on the page rather than ambient background state; also would have required per-tab state (a `Map<tabId, count>`, mirroring `resolvedContextByTab` at `background/index.ts:31`) plus two new message types with the same "tab navigated away" defensive handling the toast already needs, and introduces an orphaned-chip failure mode the badge doesn't have (chip can outlive the worker that showed it; badge cannot).
- **"Saving…" toast that morphs into the result toast** — rejected because the existing toast pipeline holds only one `currentToast` at a time and replaces it outright (`content/index.ts:158-162`); rapid saves would tear down each other's "saving" state before the user could read it, and fixing that would mean building toast aggregation just to support a feature that doesn't need per-item detail.

## Goals / Non-Goals

**Goals:**
- Give the user an ambient, low-noise signal that at least one save is currently in progress, visible without looking at the page.
- Make that signal self-correcting if the service worker dies mid-save, without adding retry/resume logic to the save flow itself.
- Cover all three save entry points (drag, context menu, snip) with one mechanism.

**Non-Goals:**
- No precision about *how many* saves are in flight, or *which* one is slow.
- No reaction to failure — outcome reporting stays the toast's and "Recent saves"'s job.
- No per-tab scoping — the badge is a single global indicator, same as the toolbar icon it lives on.
- No persistence of in-flight state across service worker restarts.

## Decisions

### 1. A single module-level counter, not a per-tab map

`extensions/src/background/index.ts` gets one `let activeSaves = 0`, incremented at the start of `handleSave`/`handleCapture` and decremented in a `finally` so it runs whether the function resolves or throws. The badge is purely `activeSaves > 0` — boolean, no count displayed.

A per-tab map was the natural structure for the chip alternative (each tab needed its own visibility), but the badge is one shared piece of toolbar chrome regardless of which tab triggered the save — a scalar is the correct shape, not an artifact of cutting corners.

### 2. Wrap the two outer entry points (`handleSave`, `handleCapture`), not `persistImage` alone

Both functions have work before `persistImage` is reached — `handleSave` does an auth check and an image fetch (`resolveImageBlob`) first. Wrapping only `persistImage` would under-report: the fetch step is real, sometimes-slow work that the badge should also cover. Wrapping the full body of `handleSave`/`handleCapture` in `try { ... } finally { activeSaves--; updateBadge(); }` (with the increment immediately on entry) covers the whole user-perceived duration, including early-return failure paths like the auth check — those just increment and decrement almost immediately, which is harmless.

### 3. Badge state is not persisted, and a fresh module load explicitly clears it

`activeSaves` lives only in the service worker's module scope. Two distinct "killed" cases behave differently here:

- **The browser itself quits.** The toolbar, the badge rendering, and the service worker are all the same process — there's nothing to clean up, the whole thing stops existing at once.
- **The service worker is idle-killed while the browser stays open.** This is the real case to design for. The `action` badge is sticky state owned by the browser UI, not something the worker has to keep redrawing — once set, it stays displayed until something explicitly clears it. If the worker dies while `activeSaves` was `> 0`, the dot stays visibly on with nothing left running to turn it off.

Relying on "the next save's own lifecycle will fix it" is not good enough — that only clears the dot once that next save *finishes*, not when it starts, so a crash-then-restart could show an uninterrupted dot across two unrelated saves with no visible gap to indicate the first one was actually abandoned. Instead, the module's top-level code (which re-runs on every cold start, alongside the existing `onInstalled` registration) calls `browser.action.setBadgeText({ text: "" })` unconditionally once, before anything else touches the badge. This guarantees every fresh worker start clears whatever was left over before any save has a chance to set it again, rather than depending on it.

A stale dot can still be visible for the window between "worker died" and "the worker's next cold start." That next cold start is driven entirely by the background script's own registered listeners — `runtime.onInstalled` (extension install/update), `runtime.onMessage` (a content script sending `drag-save`/`snip-captured`/resolved-context, from any tab), `commands.onCommand` (the snip hotkey), or `contextMenus.onClicked` — plus a full browser restart. Notably, **closing the tab that was mid-save does not wake the worker**: there is no `tabs.onRemoved` listener, so that event alone leaves the stale dot in place until one of the listeners above fires on some other tab or action. That residual window is acceptable for a purely cosmetic indicator; "Recent saves" remains the actual source of truth for what succeeded.

### 4. Remove the `DEV` badge text rather than coexist with it

`background/index.ts:11-14` sets `badgeText: "DEV"` once at module load in non-production builds. Both features want the same badge slot. Since the dev/prod distinction is a build-time convenience and not user-facing, it's dropped outright rather than adding logic to swap between "DEV" and the dot (which would also reintroduce exactly the kind of stale-state bookkeeping Decision 3 avoids).

## Risks / Trade-offs

- [Trade-off] A very fast save (small image, fast network) may show the dot for a barely-perceptible instant, or not long enough to register. → Accepted: there's no minimum-display-time requirement here; the badge is meant to reassure on *slow* saves, and a flicker on fast ones is harmless noise, not a correctness issue.
- [Trade-off] The badge can't tell the user which tab, or how many, saves are in flight. → Accepted per the Non-Goals; this is intentionally an "alive/not alive" signal, not a status dashboard.
- [Risk] A future contributor adding a fourth save entry point could forget to wrap it, silently excluding it from the indicator. → Mitigation: both `handleSave` and `handleCapture` are the only two functions that reach `persistImage`; as long as new entry points continue to funnel through one of those two (rather than calling `persistImage` directly), they're covered automatically.

## Open Questions

None outstanding — the dot-vs-count, chip-vs-badge, and failure-reaction questions were resolved during exploration prior to this proposal.
