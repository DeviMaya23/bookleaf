## 1. In-flight counter and badge helper

- [x] 1.1 In `extensions/src/background/index.ts`, add a module-level `let activeSaves = 0` and an `updateBadge()` helper that sets the badge text to a dot (`"•"`) when `activeSaves > 0` and clears it (`""`) when `activeSaves === 0`
- [x] 1.2 Unit test `updateBadge()` (or the wrapped behavior, if not exported standalone): calling it with `activeSaves === 0` clears the badge text; with `activeSaves > 0` it sets the dot

## 2. Wrap save entry points

- [x] 2.1 Wrap the full body of `handleSave` in `try { ... } finally { activeSaves--; updateBadge(); }`, incrementing `activeSaves` and calling `updateBadge()` immediately on entry, before the auth check
- [x] 2.2 Apply the same wrapping to `handleCapture`
- [x] 2.3 Unit test: `handleSave` increments the counter (badge shows the dot) before the auth check resolves, and decrements it (badge clears, if no other save is in flight) after the function settles — covering both the early-return auth-failure path and the full success path
- [x] 2.4 Unit test: `handleSave` decrements the counter on a thrown/rejected failure (image fetch error, upload error) the same as on success
- [x] 2.5 Unit test: the same increment/decrement behavior for `handleCapture`
- [x] 2.6 Unit test concurrency: two overlapping `handleSave`/`handleCapture` calls both increment the counter; the badge remains shown after only one of the two finishes, and clears only after both have finished

## 3. Cold-start badge reset and DEV badge removal

- [x] 3.1 Remove the existing `browser.action.setBadgeText({ text: "DEV" })` / `setBadgeBackgroundColor` block in non-production builds (`background/index.ts:11-14`)
- [x] 3.2 Add an unconditional `browser.action.setBadgeText({ text: "" })` call at module top-level, run before any other badge-affecting code, so every cold start clears whatever badge state was left over from a previous run
- [x] 3.3 Unit test: on module evaluation, the badge is cleared regardless of build mode (covers the removal of the `DEV` text as well as the stale-dot-clearing behavior)

## 4. Verification

- [x] 4.1 Run `npm run build` in `extensions/` and fix any errors
- [x] 4.2 Run `npm run type-check` in `extensions/` and fix any errors
- [x] 4.3 Skip `npm run lint` — no lint script exists in `extensions/` (consistent with prior extension changes)
- [x] 4.4 Manually verify: trigger a drag-to-save, a right-click save, and a snip-capture save individually — the toolbar badge shows a dot while each is in flight and clears when it finishes
- [x] 4.5 Manually verify concurrency: start a save in one tab, switch to another tab and start a second save before the first finishes — confirm the badge stays shown until both have finished, regardless of which tab is active or which finishes first
- [x] 4.6 Manually verify a failed save (e.g. disconnect network mid-upload) still clears the badge, with the existing error toast as the only outcome signal
- [x] 4.7 Manually verify non-production builds no longer show the `DEV` badge text
