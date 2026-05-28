## Context

The extension's background script (`src/background/index.ts`) is a service worker — it has no access to page DOM. All three notification call sites (`notify()`) currently use `browser.notifications`, which requires the `"notifications"` manifest permission and produces OS-level popups.

To show an in-page toast, the extension needs a **content script** running in the page's context. The background communicates with it via `browser.tabs.sendMessage`.

The frontend app uses [sonner](https://sonner.emilkowal.ski/) with `richColors`. The extension toast will visually match sonner's style but be implemented in vanilla TypeScript to avoid bundling React/Sonner into every page the user visits.

## Goals / Non-Goals

**Goals:**
- Replace all three `notify()` call sites with in-page toasts
- Toast renders inside a Shadow DOM to prevent host-page CSS interference
- Visual design matches the frontend's sonner toast (success = green, failure = red, two-line copy, rounded, bottom-right position)
- Works in both Chrome and Firefox builds

**Non-Goals:**
- Toast persistence or queuing (show one at a time is fine)
- Matching sonner's animation library exactly (a simple CSS fade/slide is sufficient)
- Showing a toast when there is no active tab (e.g., triggered from extension popup — not a scenario in scope)

## Decisions

### 1. Static content script vs. programmatic injection

**Decision:** Static content script declared in `manifest.json`.

`content_scripts` with `"matches": ["<all_urls>"]` injects the script on every page load. The alternative — `browser.scripting.executeScript` from the background — requires the `"scripting"` permission and adds async complexity (need to inject before sending the message).

Static injection is simpler and the overhead is negligible: the content script only attaches a `runtime.onMessage` listener and does nothing else until a message arrives.

### 2. Style isolation: Shadow DOM

**Decision:** Attach a Shadow Root to a dedicated host element appended to `<body>`.

Without isolation, host-page CSS resets (e.g., `* { box-sizing: border-box; margin: 0 }`, Tailwind preflight) and class name collisions will break the toast's appearance. Shadow DOM provides complete encapsulation with zero runtime cost (native browser API, no library).

```
document.body
  └── <div id="bookleaf-toast-host">  (no styles, just a mount point)
        └── Shadow Root (mode: "open")
              ├── <style>/* scoped CSS */</style>
              └── <div class="toast success|error">
                    <span class="title">...</span>
                    <span class="body">...</span>
                  </div>
```

### 3. Toast implementation: vanilla TS

**Decision:** Hand-written DOM + CSS, no React or Sonner.

The toast is two lines of text with a colored left border, icon, and fade animation — a straightforward DOM operation. Bundling React + ReactDOM + Sonner (~190KB) into the content script for this is not justified.

The CSS will mirror sonner's visual tokens: white background, `border-radius: 8px`, `box-shadow`, green/red accent.

### 4. Threading `tabId` through `handleSave`

The context menu listener has `tab?.id`. It must be forwarded into `handleSave` so the background can call `browser.tabs.sendMessage(tabId, ...)`. If `tabId` is undefined (edge case), the notification is silently dropped — this is acceptable given the context menu only fires on a real tab.

### 5. Toast copy

| Outcome | Line 1 (bold) | Line 2 |
|---|---|---|
| Success | Saved to Bookleaf. | Added to Unsorted. |
| Failure | Couldn't save image. | Check your connection and try again. |
| Not logged in | Saved to Bookleaf. | Please log in first. |

Wait — "not logged in" currently shows a toast too. Line 1 will reuse "Bookleaf" as the title for that case:

| Outcome | Line 1 (bold) | Line 2 |
|---|---|---|
| Not logged in | Bookleaf | Please log in first. |

## Risks / Trade-offs

- **Content script blocked by page CSP** → If a page's CSP blocks the content script from running, the message from the background will fail silently (`sendMessage` rejects). Wrap in try/catch; no user-visible failure beyond the missing toast.
- **Tab navigates before save completes** → `sendMessage` will reject because the content script in the old page is gone. Same mitigation: catch the rejection silently.
- **Multiple rapid saves** → If two saves complete close together, the second toast will replace the first (simplest behavior). No queueing needed for v1.
- **Firefox manifest transform** → `vite.config.ts` only transforms the `background` key; `content_scripts` passes through unchanged. Verified by inspection.

## Open Questions

None — design is complete.
