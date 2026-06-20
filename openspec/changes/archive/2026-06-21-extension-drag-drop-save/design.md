## Context

The right-click save flow has two halves that are already cleanly separated:

1. **Capture** (content script, `contextmenu` listener in `content/index.ts`): given `event.target`, walk the live DOM to resolve `srcUrl` (`resolveCardImageSrc`, card sites only) and `title` (`resolveAltText` / `resolveTweetText` / `resolveFacebookAltText`, per site), and ship the result to background via `runtime.sendMessage({ resolved })`, keyed by tab in `resolvedContextByTab`.
2. **Action** (background, `handleContextMenuClick` → `handleSave`): combines the captured context with browser-native `info.srcUrl` / `info.linkUrl`, runs `resolveHighResUrl`/`resolveHighResReferrer`/`validateCandidate` and `resolveLinkPermalink`, then fetches, thumbnails, and uploads.

A manual spike (Firefox devtools, Pinterest and Instagram) confirmed `dragstart`'s `event.target` is the same kind of element `contextmenu`'s is — the `<a>` card wrapper, not the `<img>` — so the existing capture-side resolvers apply unchanged. The action side needs no changes at all: it already accepts plain strings, not browser-native menu-click objects, for the parts a drag would supply (`handleSave` itself takes `{ srcUrl, pageUrl, title, tabId }`).

One real gap exists on the capture side: `contextmenu`'s non-card path never needed a content-script-side `srcUrl` resolver, because the browser supplies `info.srcUrl` natively whenever the right-click context is `"image"`. Drag has no equivalent native signal — the content script must determine `srcUrl` itself even on plain (non-card) sites.

## Goals / Non-Goals

**Goals:**
- Reuse `resolveCardImageSrc`, `resolveAltText`/`resolveTweetText`/`resolveFacebookAltText`, `closest("a")?.href`, `resolveLinkPermalink`, `resolveHighResUrl`/`resolveHighResReferrer`/`validateCandidate`, and `handleSave` completely unchanged.
- Add a `dragstart`-time capture path in content script that produces the same `{ srcUrl, title, linkUrl }` shape the `contextmenu` path produces, for both card sites and plain (`<img>`-only) sites.
- Render a transient drop-zone anchored near the `dragstart` pointer position, visible only between `dragstart` and `dragend`, and only when a valid `srcUrl` was resolved.
- On `drop` inside the zone, send one message to background with the captured context; background calls `handleSave` directly.

**Non-Goals:**
- No `dataTransfer` parsing (no `text/html`/`srcset` extraction) — superseded by live-DOM capture at `dragstart`.
- No fix to the `linkPermalinkRules.ts` Pinterest/Instagram gap — drag intentionally inherits the same `sourceUrl` imprecision right-click has on those sites today, to keep both paths consistent. Tracked as a separate, optional follow-up.
- No persistent/hover-based save UI (explicitly rejected in favor of a gesture-triggered, transient affordance).
- No support for sites that disable native image dragging (`draggable="false"` / `-webkit-user-drag: none`) — right-click remains the fallback there.

## Decisions

**1. Capture context at `dragstart`, not at `drop`, and hold it in a module-level variable in content script (no `runtime.sendMessage` round-trip needed for the capture itself).**
Unlike `contextmenu` → `contextMenus.onClicked`, which crosses a browser-native UI boundary into background (hence the `resolvedContextByTab` map), `dragstart` and `drop` both fire inside the same content script, on the same page. A plain in-memory variable, set on `dragstart` and read on `drop`, is sufficient — simpler than threading it through a tab-keyed map for no reason.

**2. `srcUrl` resolution adds two generic fallback tiers beneath the existing site-specific resolver.**
```
resolveDragImageSrc(target, pageUrl):
  if shouldResolveCardDom(pageUrl):
    cardSrc = resolveCardImageSrc(target)
    if cardSrc: return cardSrc                                          // unchanged, but now falls through on null
  if target.tagName === "IMG": return target.src                        // generic tier
  descendantImg = target.querySelector("img")
  if descendantImg: return descendantImg.src                           // descendant tier
  return null
```
The plain-`<img>` tier is the direct analogue of what the browser already does for free on the `contextmenu` path (telling you `info.srcUrl` when you right-click an `<img>`) — it's not new site-specific logic, just the one-line equivalent for drag.

Two refinements were added after manual cross-site testing surfaced real gaps in the originally planned 2-tier version:
- **Card-site resolution now falls through instead of short-circuiting to `null`.** The original version treated "card site" and "plain `<img>`" as mutually exclusive, so on a card site (e.g. Pinterest), any drag where `resolveCardImageSrc` found no enclosing `<a>`-wrapped card (e.g. the full-image/lightbox view, which isn't wrapped in the card's link structure) resolved to `null` with no drop zone — even though the dragged element was a plain `<img>`. Falling through fixes this without weakening the card-site path itself.
- **Descendant-`<img>` tier added for non-`<img>` drag targets.** On Twitter and Facebook, the actual native drag target turned out to be a wrapping element (likely one explicitly marked `draggable="true"` by the site's own UI code) rather than the `<img>` itself, so the plain-`<img>` tier never matched even though an `<img>` was present as a child. Searching `target.querySelector("img")` as a last resort catches this case.

Sites with neither a card rule, a direct `<img>` under the drag, nor an `<img>` descendant (e.g. CSS `background-image` cards) still resolve to `null` and simply don't get a drop zone, same as today's right-click behavior on such sites (no menu item appears).

**3. `linkUrl`/`sourceUrl` is captured the same way as `srcUrl` — via the live DOM (`target.closest("a")?.href`), then run through the unchanged `resolveLinkPermalink` gate — not read from `dataTransfer`'s `text/uri-list`.**
This was the key correction from the original spike-based plan: `dataTransfer`'s permalink bypasses `resolveLinkPermalink` entirely, which would make drag silently more precise than right-click on Pinterest/Instagram (where that gate doesn't yet recognize those hosts). Gating drag's permalink through the same function keeps both paths behaviorally identical, modulo trigger gesture.

**4. The drop zone is a fixed-position element rendered into the existing toast shadow host (`#bookleaf-toast-host`), positioned at the `dragstart` client coordinates, created on `dragstart` and unconditionally removed on `dragend` — plus `pointerup`/`mouseup` as a safety net.**
Reusing the existing shadow-DOM host avoids introducing a second injection point. Cleanup on `dragend` (not `drop`) ensures the zone disappears even if the user cancels the drag (drops outside any valid target, presses Escape, etc.) — `dragend` always fires after `dragstart` for *native* drags, `drop` does not.

Manual testing surfaced a case the original `dragend`-only plan didn't cover: sites with their own drag-and-drop UI (e.g. Bookleaf's own app, which uses dnd-kit) commonly call `preventDefault()` on `dragstart` to cancel the browser's native drag operation in favor of synthetic pointer-based dragging. When native drag is cancelled this way, `dragend` never fires at all, leaving the drop zone stuck on screen indefinitely. `pointerup` and `mouseup` always fire once the user releases the mouse button, regardless of whether a native drag was ever actually in progress, so adding them as additional cleanup triggers closes this gap. This doesn't fire spuriously during a real native drag-drop, since native HTML5 DnD suppresses ordinary mouse events for the duration of the drag (the browser fires `drop`/`dragend` on release instead).

**5. The drop zone only renders when `dragstart` resolved a non-null `srcUrl`.**
Prevents visual noise for unrelated drags (selected text, non-image links, etc.). The check is synchronous and cheap (same resolvers already used for right-click), so it can run inside the `dragstart` handler itself before deciding whether to render anything.

**6. On `drop`, content script sends one message (e.g. `{ type: "drag-save", srcUrl, title, sourceUrl, pageUrl }`) to background; background's handler calls `handleSave` directly — no new `resolveTitle`/permalink logic in background.**
Mirrors `handleContextMenuClick`'s shape but skips the `resolvedContextByTab` lookup and `info`-object plumbing, since content script already has everything resolved at message-send time.

**7. Drag-and-drop save is disabled outright on Bookleaf's own app, gated by origin match against `VITE_APP_URL`.**
The content script is injected via `<all_urls>`, which includes Bookleaf's own web app. Manual testing showed the drag-save trigger firing on Bookleaf's own images, which makes no sense (the image is already in Bookleaf) and additionally interacts badly with Bookleaf's own dnd-kit-based drag UI (see Decision 4's `pointerup`/`mouseup` safety net, which was found via this same testing). The check is a simple origin comparison (`new URL(pageUrl).origin === new URL(import.meta.env.VITE_APP_URL).origin`) performed at the top of the `dragstart` handler, before any resolution work.

## Risks / Trade-offs

- **[Risk] Empirical dependence on `dragstart` landing on the same element `contextmenu` does.** Verified on Pinterest and Instagram only. → **Mitigation:** the resolver functions are identical to right-click's, so any site working for right-click's card path works for drag too; a site needing card support for drag but not yet covered just needs a `cardDomResolveRules` entry, the same maintenance burden as today.
- **[Risk] Some sites disable native image dragging entirely** (no `dragstart` fires). → **Mitigation:** graceful no-op — no drop zone appears, right-click remains available, no regression. Confirmed during manual testing on Instagram's full-image view, which appears not to support native dragging at all.
- **[Risk] Drop zone could visually collide with a site's own drag-and-drop UI** (e.g. drag-to-reorder). → **Confirmed in manual testing**, on a third-party site with its own drag-and-drop interactions (outside the explicitly tested site list) — the zone's `dragover`/`drop` interception did cause minor visual/behavioral interference with that site's own UI. **Accepted as-is**: the site in question is not a primary target for this extension, the interference doesn't break the site's functionality, and the zone is already kept small/anchored (not full-page) per the original mitigation. Not fixed further in this change.
- **[Trade-off] Pinterest/Instagram `sourceUrl` stays imprecise on both paths.** Accepted deliberately (see Non-Goals) to avoid silent behavioral drift between right-click and drag; a follow-up change can add those hosts to `linkPermalinkRules.ts` and fix both paths at once.

## Migration Plan

Purely additive — no data model, no persisted state, no flag. Ships as new listeners in `content/index.ts` and a new message handler in `background/index.ts`; right-click path is untouched. Rollback is a plain revert.

## Open Questions

- Exact visual treatment of the drop zone (size, icon, copy, offset from pointer) — implementation detail, not architectural; can be decided during build.
- Whether to eventually detect "drag started but no `dragstart` ever fires" (sites that disable native dragging) and surface a hint to use right-click instead — out of scope for this change, noted as a possible future enhancement.
