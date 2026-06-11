## ADDED Requirements

### Requirement: Right panel field-display sections are split into focused presentational components
`RightPanel.tsx`'s image details summary and download action SHALL each be extracted into their own presentational component under `frontend/src/features/right-panel/components/`, rather than inline within `RightPanel.tsx`.

#### Scenario: Details grid component exists
- **WHEN** the codebase is inspected
- **THEN** `frontend/src/features/right-panel/components/DetailsGrid.tsx`
  exists

#### Scenario: Download button component exists
- **WHEN** the codebase is inspected
- **THEN** `frontend/src/features/right-panel/components/DownloadButton.tsx`
  exists

### Requirement: Folder and tag chip-inputs share a single token-input implementation
`FolderInput` and `TagInput` SHALL be implemented as configurations of a shared `TokenInput` component under `frontend/src/features/right-panel/components/`, rather than duplicating chip-input, dropdown, and keyboard-navigation logic across two files.

#### Scenario: TokenInput component exists
- **WHEN** the codebase is inspected
- **THEN** `frontend/src/features/right-panel/components/TokenInput.tsx`
  exists

### Requirement: Vision-suggestion logic lives under app-shell/
`useVisionSuggestion`, which powers `AppLayout`'s post-upload "suggested folder" toast, SHALL reside under `frontend/src/app-shell/`, not under `frontend/src/features/right-panel/hooks/`, since it has no relationship to right-panel/image-detail editing.

#### Scenario: useVisionSuggestion lives under app-shell
- **WHEN** the codebase is inspected
- **THEN** `frontend/src/app-shell/useVisionSuggestion.ts` exists and
  `frontend/src/features/right-panel/hooks/useVisionSuggestion.ts` does not
