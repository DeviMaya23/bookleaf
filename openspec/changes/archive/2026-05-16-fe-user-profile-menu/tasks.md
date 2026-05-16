## 1. Frontend — ProfileMenu Component

- [x] 1.1 Create `src/components/ProfileMenu.tsx` — fetch user profile via `useKindeAuth().getUserProfile()` on mount
- [x] 1.2 Render avatar using shadcn `Avatar` with `AvatarImage` (picture URL) and `AvatarFallback` (initials)
- [x] 1.3 Wrap trigger and dropdown in shadcn `DropdownMenu` with a single Sign out item that calls `logout()`
- [x] 1.4 Add `ProfileMenu` to the footer of `FolderSidebar`, below the folder list
- [x] 1.5 Remove `LogoutButton` component and all imports/usages

## 2. Frontend — Tests

- [x] 2.1 Write unit tests for `ProfileMenu`: renders avatar with picture, renders initials fallback, Sign out calls `logout()`
