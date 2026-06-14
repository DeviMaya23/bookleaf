## Context

`extensions/vite.config.ts` transforms the manifest for any Firefox-like mode (`firefox`, `firefox-production`) and currently injects a single hardcoded `browser_specific_settings.gecko.id: "bookleaf@evimay.me"`. Both the dev build (`build:firefox`) and the signed prod build (`build:firefox:prod`, signed via `npm run sign:firefox`) use this same ID, so Firefox treats them as one add-on — installing the signed prod `.xpi` displaces a temporarily-loaded dev build on the same profile, and both share one `browser.identity.getRedirectURL()` value (and therefore one Kinde redirect URI).

## Goals / Non-Goals

**Goals:**
- Dev (`build:firefox`) and prod (`build:firefox:prod`) Firefox builds use different gecko IDs so both can be installed on the same Firefox profile without conflict.
- Prod keeps the existing ID (`bookleaf@evimay.me`) so the already-registered Kinde redirect URI for prod keeps working.

**Non-Goals:**
- No change to the Chrome build or its manifest transform.
- No code changes to register the new dev redirect URI in Kinde — that's a manual dashboard step.
- No change to the signing/packaging flow itself.

## Decisions

- In `vite.config.ts`, select the gecko ID from the `mode` already available in the `defineConfig(({ mode }) => ...)` callback: `mode === "firefox-production" ? "bookleaf@evimay.me" : "bookleaf-dev@evimay.me"`. This keeps the existing single `isFirefox` branch and `transformManifest` function, just parameterizing the one literal.
  - Alternative considered: separate `vite.firefox.config.ts` / `vite.firefox-production.config.ts` files — rejected as overkill for a single string value and would duplicate the rest of the shared config.

## Risks / Trade-offs

- [Dev OAuth login breaks until the new dev gecko ID's redirect URI is registered in the dev Kinde app] → Mitigation: documented as a manual follow-up step in tasks; existing prod redirect URI is untouched so prod is unaffected.
- [Anyone with the old dev build installed will see a "new" add-on appear after rebuilding rather than an update to the existing one] → Mitigation: one-time manual removal of the old dev install; low impact since this is a personal/dev-only profile.

## Migration Plan

1. Update `vite.config.ts` gecko ID logic.
2. Rebuild the dev Firefox extension (`npm run build:firefox`) and reload it in `about:debugging` — Firefox will register it as a new add-on (`bookleaf-dev@evimay.me`).
3. Register the new dev redirect URI (`https://<new-extension-id>.extensions.allizom.org/`, derived from `bookleaf-dev@evimay.me`) as an allowed callback in the dev Kinde app.
4. Remove the old temporarily-loaded dev add-on if it lingers under the previous shared ID.

No rollback complexity — reverting `vite.config.ts` restores the previous shared-ID behavior.
