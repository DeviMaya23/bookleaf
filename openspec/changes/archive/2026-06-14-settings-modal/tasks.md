## 1. Shared primitives & lib

- [x] 1.1 Add `components/ui/switch.tsx`, wrapping `@base-ui/react/switch` following the existing `toggle.tsx`/`dialog.tsx` thin-wrapper pattern.
- [x] 1.2 Add `deleteMe(getToken)` to `features/auth/lib/me.ts`, calling `apiFetch('/me', getToken, { method: 'DELETE' })`.

## 2. SettingsModal shell

- [x] 2.1 Create `features/settings/` directory with `SettingsModal.tsx`: left-nav with Account/App/Advanced sections, `useState` for active section (default `'account'`), renders selected section's content.
- [x] 2.2 Use `Dialog`/`DialogContent` from `components/ui/dialog.tsx`, overriding width via `className` (e.g. `sm:max-w-2xl`) on this instance only; build the left-nav + content layout with plain `div`s, matching the design handoff (fonts, backgrounds, borders).
- [x] 2.3 Add unit tests for `SettingsModal`: default section is Account, switching sections updates displayed content and active nav highlight, close control closes the modal.

## 3. Wire up ProfileMenu

- [x] 3.1 Add a "Settings" `DropdownMenuItem` to `ProfileMenu.tsx`, placed above the divider/"Sign out" item.
- [x] 3.2 Add `open`/`onOpenChange` state in `ProfileMenu.tsx` for `SettingsModal`; clicking "Settings" opens it.
- [x] 3.3 Update `ProfileMenu.test.tsx` for the new dropdown item and modal-open behavior.

## 4. Account section — profile & log out

- [x] 4.1 Create `AccountSection.tsx`: read-only profile display (avatar/initials fallback, full name, email) from the existing `['profile']` query (`getUserProfile()`), reusing `getInitials`/`getFullName` helpers from `ProfileMenu.tsx`.
- [x] 4.2 Add a "Log out" action calling `logout()` from `useKindeAuth()`.
- [x] 4.3 Add unit tests: profile fields render read-only, initials fallback when no `picture`, "Log out" calls `logout()`.

## 5. App section — theme display

- [x] 5.1 Create `AppSection.tsx`: call `useTheme()` and render a single non-interactive "Default — Warm parchment" row shown as selected.
- [x] 5.2 Add unit test: the Default theme row renders as selected, reflecting `useTheme()`'s value.

## 6. Advanced section — AI features toggle

- [x] 6.1 Create `AdvancedSection.tsx`: read `useQuery(['me'], () => getMe(getToken))` and render `<Switch checked={me?.vision_enabled ?? false} disabled />` with an "AI features" label.
- [x] 6.2 Add unit tests: toggle renders on/disabled when `vision_enabled` is `true`, off/disabled when `false`.

## 7. Delete account flow

- [x] 7.1 In `AccountSection.tsx`, add the "Delete account" entry point and a confirmation view with a text input for the account email.
- [x] 7.2 Disable the confirm action until the typed value matches `profile.email` (trimmed, case-insensitive).
- [x] 7.3 Wire a `useMutation` calling `deleteMe(getToken)`.
- [x] 7.4 On success: call `logout()` and navigate to `/`. Verify the ordering against `useKindeAuth().logout()`'s actual behavior (immediate redirect vs. local session clear) and sequence accordingly per design.md's noted risk.
- [x] 7.5 On error: show a toast via `sonner` describing the failure, remain on the confirmation view with the typed email preserved.
- [x] 7.6 Add unit tests: confirm disabled/enabled based on email match, successful deletion triggers `logout()` + navigation to `/`, failed deletion shows a toast and preserves the confirmation view and typed email.

## 8. Verification

- [x] 8.1 Run `npm run build` and fix any errors.
- [x] 8.2 Run `npm run lint` and fix any issues.
