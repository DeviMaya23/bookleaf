## Why

The extension scaffold enables login but has no actual saving functionality. This change delivers the core user-facing feature: right-clicking any image on any webpage and saving it to Bookleaf in one action, without leaving the browser.

## What Changes

- New context menu item "Save to Bookleaf" appears when right-clicking any `<img>` element
- Background service worker fetches the image blob directly (CORS bypassed via `host_permissions: ["<all_urls>"]`)
- Reuses the existing 3-step upload flow: `POST /images` → `PUT` to R2 presigned URL → `POST /images/:id/complete`
- Image title defaults to the current page title; `source_url` is set to the page URL
- Image always saves to root (no folder selection)
- Chrome notification shown on save success or failure
- `contextMenus` and `notifications` permissions added to the extension manifest

## Capabilities

### New Capabilities

- `extension-save-image`: Context menu registration, image fetch via service worker, 3-step upload orchestration, and Chrome notification feedback

### Modified Capabilities

- `extension-scaffold`: Manifest gains `contextMenus` and `notifications` permissions; background service worker gains the save-image logic

## Impact

- No backend changes — reuses existing `POST /images`, presigned R2 PUT, and `POST /images/:id/complete` endpoints
- Extension only: changes live entirely in `/extensions/src/`
- Requires the user to be logged in; unauthenticated save attempts show a "Please log in first" notification
