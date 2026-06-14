## Why

The Firefox build currently injects a single hardcoded `browser_specific_settings.gecko.id` (`bookleaf@evimay.me`) regardless of build mode. This means the dev build (loaded temporarily via `about:debugging`) and the signed production build (self-distributed `.xpi`) are treated by Firefox as the same add-on, so installing one displaces the other on the same profile, and both share a single OAuth redirect URI. Splitting the gecko ID by build mode lets dev and prod builds coexist on the same Firefox profile.

## What Changes

- `vite.config.ts`'s Firefox manifest transform sets `browser_specific_settings.gecko.id` based on build mode:
  - `firefox` (dev) → `bookleaf-dev@evimay.me`
  - `firefox-production` (prod) → `bookleaf@evimay.me` (unchanged)
- Update the `extension-firefox-compat` spec to document the mode-dependent gecko ID instead of a single fixed value.

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

- `extension-firefox-compat`: the gecko ID requirement changes from a single hardcoded value for "the Firefox build" to a mode-dependent value (dev vs. production builds get different IDs).

## Impact

- `extensions/vite.config.ts` — Firefox manifest transform logic
- `openspec/specs/extension-firefox-compat/spec.md` — requirement and scenario updates
- Dev builds will get a new gecko ID (`bookleaf-dev@evimay.me`), which means a separate OAuth redirect URI must be registered against the dev Kinde app (follow-up manual step, not part of this change's code)
