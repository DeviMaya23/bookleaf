## REMOVED Requirements

### Requirement: Gallery self-polls while any image has a pending thumbnail
**Reason**: All upload paths (web app and browser extension) now set `thumbnail_url` synchronously at upload completion. No image will ever have `thumbnail_url === null` after being created, so the polling interval is permanently inactive dead code.
**Migration**: Remove the `refetchInterval` option from the `useInfiniteQuery` call in `ImageGrid.tsx`. The default behaviour (no polling) is correct.
