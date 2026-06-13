## Context

`ProfileMenu.tsx` (in `features/auth/components/`) currently renders the user's avatar/name (via `useQuery(['profile'], () => getUserProfile())`, `staleTime: Infinity`) and a single "Sign out" item. The backend already exposes `DELETE /me` (full account wipe, `account-deletion` spec, merged) and `GET /me` (returns `vision_enabled`, consumed by `useVisionSuggestion` via `useQuery(['me'], () => getMe(getToken))`). The `fe-theme` capability provides `useTheme()` / `ThemeProvider`, currently supporting a single theme value (`Theme = 'warm'`). The shared `Dialog`/`DialogContent` primitives (`components/ui/dialog.tsx`, Base UI) default to `sm:max-w-sm`, sized for small confirmation dialogs.

## Goals / Non-Goals

**Goals:**
- Add a "Settings" entry to `ProfileMenu` that opens a `SettingsModal` with Account / App / Advanced sections.
- Wire the Account section's profile display to the existing `['profile']` query (no new request).
- Wire the App section's theme row to the real `useTheme()` hook.
- Wire the Advanced section's AI toggle to the existing `['me']` query (`vision_enabled`), read-only.
- Implement the delete-account flow end-to-end against the existing `DELETE /me` endpoint, with email-match confirmation, toast on error, logout + redirect to `/` on success.

**Non-Goals:**
- Connected accounts (Google/GitHub linking) — not built, not stubbed.
- A real theme picker (Light/Dark) — only the existing single "warm" theme is displayed.
- Making the AI features toggle interactive — no `PATCH /me` exists; toggle stays disabled.
- Editable display name / profile fields — Account section is read-only.

## Decisions

**1. New `features/settings/` directory for modal + sections.**
`SettingsModal.tsx` (shell, section nav, active-section state), `AccountSection.tsx`, `AppSection.tsx`, `AdvancedSection.tsx`. `ProfileMenu.tsx` gains a "Settings" `DropdownMenuItem` (placed above the existing divider/"Sign out", matching the design handoff's ordering) and owns the `open`/`onOpenChange` state passed to `<SettingsModal />`.

**2. Reuse `Dialog`/`DialogContent`, override width locally.**
The shared `DialogContent` defaults to `sm:max-w-sm`. Rather than changing that default (which would affect every other dialog in the app), `SettingsModal` passes a wider `className` (e.g. `sm:max-w-2xl`) on its own `DialogContent` instance, with an internal flex layout (left nav rail + content pane) built from plain `div`s — `DialogHeader`/`DialogFooter` aren't used since the layout doesn't match their stacked structure.

**3. Section switching is local state, not routes.**
`SettingsModal` holds `useState<'account' | 'app' | 'advanced'>('account')`. No URL/route changes — the modal is ephemeral UI, consistent with other modals in the app (e.g. upload modal).

**4. Profile data comes from the existing `['profile']` query.**
`AccountSection` calls the same `useQuery(['profile'], () => getUserProfile())` (`staleTime: Infinity`) that `ProfileMenu` already populates — opening the modal from the menu is guaranteed cache-warm, no extra request. Avatar/name/email rendering mirrors `ProfileMenu`'s existing `getInitials`/`getFullName` helpers.

**5. App section reads `useTheme()` but doesn't offer a picker yet.**
`AppSection` calls `useTheme()` and renders a single non-interactive row ("Default — Warm parchment", selected) using the real `theme` value. No `setTheme` call. When additional themes are added to `fe-theme`, this becomes the natural place to add a picker — the hook wiring is already in place.

**6. New `Switch` UI primitive for the AI toggle.**
`@base-ui/react` is already a dependency; `components/ui/switch.tsx` wraps `@base-ui/react/switch` following the same thin-wrapper pattern as `toggle.tsx`/`dialog.tsx`. `AdvancedSection` reads `useQuery(['me'], () => getMe(getToken))` (same key as `useVisionSuggestion`, cache-shared) and renders `<Switch checked={me?.vision_enabled ?? false} disabled />`.

**7. Delete-account flow.**
- New `deleteMe(getToken)` in `features/auth/lib/me.ts`, calling `apiFetch('/me', getToken, { method: 'DELETE' })`.
- `AccountSection` holds a `useMutation` wrapping `deleteMe`.
- Confirmation UI: a text input where the user types their account email. The "Delete account" action is disabled until the trimmed, lowercase input matches the trimmed, lowercase `profile.email`.
- On success: call `logout()` (Kinde) and `navigate('/')`.
- On error: `toast.error(...)` (sonner, already a dependency), remain on the confirmation view with the typed email preserved.

## Risks / Trade-offs

- [Risk] Overriding `DialogContent`'s width per-instance could drift from the shared dialog style → Mitigation: override via `className` on this instance only; don't change `dialog.tsx` defaults.
- [Risk] Email-match confirmation could false-negative on casing/whitespace differences from Kinde → Mitigation: trim + lowercase both sides before comparing.
- [Risk] Ordering of `logout()` and `navigate('/')` — if Kinde's `logout()` performs an immediate redirect, a subsequent `navigate('/')` may not run → Mitigation: verify `useKindeAuth().logout()` behavior during implementation and sequence the redirect accordingly (e.g. pass a post-logout redirect target if the SDK supports it, or navigate first and logout after).
- [Risk] New `Switch` component adds a UI primitive not yet used in the app → Low risk: `@base-ui/react` is already a dependency, and the wrapper follows the existing pattern used by `toggle.tsx`/`dialog.tsx`.
