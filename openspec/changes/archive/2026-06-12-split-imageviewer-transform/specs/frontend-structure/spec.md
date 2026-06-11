## ADDED Requirements

### Requirement: Image viewer transform engine is a focused hook under features/viewer/hooks/
`ImageViewer.tsx`'s pan/zoom/rotate/flip transform engine (state, fit/resize calculations, wheel and drag-to-pan handling) SHALL be extracted into a named hook under `frontend/src/features/viewer/hooks/`, with a colocated test file, rather than bundled into the component alongside its toolbar and viewport rendering.

#### Scenario: Image transform hook exists
- **WHEN** the codebase is inspected
- **THEN** `frontend/src/features/viewer/hooks/useImageTransform.ts` and
  `frontend/src/features/viewer/hooks/useImageTransform.test.ts` exist
