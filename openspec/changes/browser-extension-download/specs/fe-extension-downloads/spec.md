## ADDED Requirements

### Requirement: Extension download URL constants

The app SHALL provide a `src/lib/downloads.ts` module exporting two named string constants: `EXTENSION_FIREFOX_URL` (the R2-hosted URL for the Firefox `.xpi` file) and `EXTENSION_CHROME_URL` (the R2-hosted URL for the Chrome `.zip` file). These constants SHALL be the single source of truth for extension download URLs, used by both the Extensions public page and the Settings modal Extensions section.

#### Scenario: Firefox URL constant is defined and non-empty

- **WHEN** `EXTENSION_FIREFOX_URL` is imported from `src/lib/downloads.ts`
- **THEN** it is a non-empty string

#### Scenario: Chrome URL constant is defined and non-empty

- **WHEN** `EXTENSION_CHROME_URL` is imported from `src/lib/downloads.ts`
- **THEN** it is a non-empty string
