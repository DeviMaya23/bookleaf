## REMOVED Requirements

### Requirement: Logout button

**Reason**: Replaced by the `ProfileMenu` component, which incorporates sign-out as a dropdown action alongside user identity display.
**Migration**: Remove the `LogoutButton` component and any references to it in `AppLayout` or other consumers. Sign-out is now triggered via the Sign out item in `ProfileMenu`.
