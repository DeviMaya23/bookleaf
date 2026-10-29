# FE Audit — app-shell · lib · hooks · components

**Scope**: `frontend/src/app-shell`, `frontend/src/lib`, `frontend/src/hooks`, `frontend/src/components`  
**Files reviewed**: 55 files across 4 directories

## Summary

The shared layer and app-shell are fundamentally healthy: Base UI is used correctly throughout, no Radix imports, query-key discipline is solid, and the three dedicated hooks already extracted from `AppLayout` (`useAppView`, `useAppDragAndDrop`, `useSSEEvents`) are clean and well-tested. The two largest problems are both in `AppLayout.tsx`: it has grown to 508 lines by keeping four extractable concerns inline, and a mandatory `react-hooks/exhaustive-deps` explanation comment is missing. `lib/images.ts` has a copy-paste duplication that will cause drift. Two shared hooks also lack colocated tests.

---

## Patterns

### Debt — Multiple inline concerns in AppLayout (4 concerns, 1 file)

`frontend/src/app-shell/AppLayout.tsx` — ~300 of 508 lines belong to four concerns that are each describable in one sentence without mentioning the others:

| Concern | State pieces | Handlers |
|---------|-------------|---------|
| File drop-to-upload | `isFileDragOver`, `isAutoUploading` | `handleMainDragOver`, `handleMainDragLeave`, `handleMainDrop` |
| Upload modal triggers | `uploadOpen`, `uploadInitialFile`, `batchUploadOpen`, `batchInitialFiles` | paste `useEffect` (line 191), open/close handlers |
| Selection mode | `selectMode`, `selectedIds`, `mainSelectedId`, `bulkTrashMutation` | `handleSelectModeToggle`, `exitSelectMode`, `handleApplySelection`, `handleMoveSelectionToTrash`, `handleDownloadSelection`, `handleSelectionChange` |
| Viewer/lightbox | `selectedImage`, `autoFocusTitle`, `viewerImage`, `lightboxImage` | `handleImageSelect`, `handleViewDetails`, `handleImageDoubleClick`, `handleImageDeleted` |

The project rule: "One hook per cohesive concern (state + its effects + its handlers)." Each of these could be a hook (`useFileDrop`, `useUploadTriggers`, `useSelectionMode`, `useViewerState`). Today the file is navigable; after the next feature it will not be.

*Go analogy*: this is a 500-line handler function in a service layer — each section is logically independent but entangled by proximity.

Occurrences: `AppLayout.tsx:52–342` (all four concerns)

---

### Debt — `getImages` / `getAllImages` copy-paste (2 occurrences)

`frontend/src/lib/images.ts:36` (`getImages`) and `frontend/src/lib/images.ts:77` (`getAllImages`) share an identical 9-parameter body. The only difference is that `getImages` sets `params.set('unfiled', 'true')`. Adding, removing, or renaming a filter parameter requires two edits; one is easy to miss.

The fix is a single internal helper:
```ts
async function fetchImages(getToken, params, query) { ... }
```
with `getImages` passing `unfiled=true` and `getAllImages` not passing it.

*Go analogy*: two handler functions that copy-paste the same middleware chain, differing by one middleware.

Occurrences: `images.ts:36`, `images.ts:77`

---

### Debt — `eslint-disable` missing explanation comment (1 occurrence)

`frontend/src/app-shell/AppLayout.tsx:235`:

```ts
// eslint-disable-next-line react-hooks/exhaustive-deps
}, [location.state])
```

`CONVENTIONS.md` requires: "add a comment explaining why when using these." The omitted deps (`gallery.setFilterTagIds`, `routerNavigate`) are stable references — intentionally excluded so the effect only fires when navigation state changes, not on every render. Without that explanation, a future developer will not know whether the omission was deliberate and may reintroduce the missing deps, causing the filter to re-apply on unrelated re-renders.

Occurrence: `AppLayout.tsx:235`

---

### Debt — Hooks without colocated test files (2 hooks)

`frontend/src/hooks/useDebouncedValue.ts` and `frontend/src/hooks/usePersistedBoolean.ts` both have behavioral state and effects but no `*.test.ts` sibling. The project rule requires a colocated test for any hook. `useIsCoarsePointer` and `useTheme` both comply; these two do not.

- `useDebouncedValue`: two meaningful scenarios — value debounces after the delay, timer resets when value changes before firing.
- `usePersistedBoolean`: reads from and writes to `localStorage`; the localStorage error path (`try/catch`) is untested.

Occurrences: `hooks/useDebouncedValue.ts:1`, `hooks/usePersistedBoolean.ts:1`

---

### Debt — `isAutoUploading` local state duplicates mutation isPending (1 occurrence)

`frontend/src/app-shell/AppLayout.tsx:60,306–342`:

```ts
const [isAutoUploading, setIsAutoUploading] = useState(false)
// ...
setIsAutoUploading(true)
try { ... } finally { setIsAutoUploading(false) }
```

The rubric flags "local `isSubmitting`-style state that duplicates a mutation's `isPending`." A `useMutation` for the file auto-upload would give `isPending` for free, centralize the error path, and naturally belong in an extracted `useFileDrop` hook.

Occurrence: `AppLayout.tsx:60`

---

### Debt — Paste listener re-subscribes on every `uploadOpen` toggle (1 occurrence)

`frontend/src/app-shell/AppLayout.tsx:191–206`:

```ts
useEffect(() => {
  function handlePaste(event: ClipboardEvent) {
    if (uploadOpen) return   // reads uploadOpen from closure
    // ...
  }
  document.addEventListener('paste', handlePaste)
  return () => document.removeEventListener('paste', handlePaste)
}, [uploadOpen])             // re-attaches on every toggle
```

Every `uploadOpen` toggle removes and immediately re-adds the listener — a brief window where two listeners could fire between the remove and add, and a remove/add on every modal open/close is unnecessary overhead. The project already uses the "ref-mirrors-latest-state" pattern legitimately (`CONVENTIONS.md`); mirroring `uploadOpen` in a ref and using `[]` deps would make this a one-time subscription.

Occurrence: `AppLayout.tsx:191`

---

### Style — `"use client"` directive in three Vite component files (3 occurrences)

`"use client"` at the top of `components/ui/dialog.tsx:1`, `components/ui/dropdown-menu.tsx:1`, and `components/ui/switch.tsx:1` is a Next.js RSC directive. In Vite it is a no-op string expression. Harmless but misleading: a new developer reading these files may assume this is a Next.js project.

Occurrences: `dialog.tsx:1`, `dropdown-menu.tsx:1`, `switch.tsx:1`

---

## Unjustified hooks and state

All `useEffect`, `useState`, `useMemo`, `useCallback`, and `useRef` instances in scope are justified. Specific notes:

- `AppLayout.tsx:73–85` — two separate reset effects on view change (`viewKey` and `view.type`). Both are legitimate navigation-driven clears. The second could be merged into the first by checking `view.type` inside one effect, but this is cosmetic.
- `AppLayout.tsx:228–236` — reads navigation state from `location.state` and clears it; valid one-shot effect, but see the missing comment finding above.
- `useSSEEvents.ts:19–71` — legitimate external sync (SSE connection lifecycle).
- `useTheme.tsx:22–25` — legitimate external sync (DOM dataset + localStorage).
- `useIsCoarsePointer.ts:8–15` — legitimate external sync (MediaQueryList listener).
- `useDebouncedValue.ts:6–9` — legitimate external sync (setTimeout).

---

## Suggested fix order

1. **Extract the four inline concerns from AppLayout** — removes the main growth blocker; each extracted hook is easier to test in isolation and each future feature can touch a file of ≤ 80 lines instead of navigating 500.
2. **Deduplicate `getImages`/`getAllImages`** — any new filter parameter currently requires two edits; fix this before the next filter is added.
3. **Add the eslint-disable explanation comment** — takes 30 seconds; without it a future author will re-introduce the deps and cause silent behavior change.
4. **Add test files for `useDebouncedValue` and `usePersistedBoolean`** — the `localStorage` error path in `usePersistedBoolean` is completely uncovered.
5. **Replace `isAutoUploading` with a mutation** — natural follow-on after extracting `useFileDrop`; low urgency on its own.

---

## Possible new rules for CONVENTIONS.md

- *eslint-disable comment format*: "When adding an `eslint-disable-next-line react-hooks/exhaustive-deps` comment, the line immediately above it must explain in plain prose which deps are excluded and why (e.g., `// gallery.setFilterTagIds is stable; fire only on state change, not every render`)." The current rule says "needs a comment explaining why" but no format is specified, leading to the missing-comment pattern seen here.
