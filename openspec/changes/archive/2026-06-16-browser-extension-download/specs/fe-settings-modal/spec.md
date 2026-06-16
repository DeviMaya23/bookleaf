## MODIFIED Requirements

### Requirement: SettingsModal shell and navigation

The system SHALL provide a `SettingsModal` component, opened via the `ProfileMenu`'s "Settings" item. The modal SHALL display a left-hand navigation listing four sections — **Account**, **App**, **Advanced**, and **Extensions** — and SHALL render the content of the currently selected section. The **Account** section SHALL be selected by default when the modal opens. The modal SHALL be dismissible via a close control.

#### Scenario: Modal opens to Account section by default

- **WHEN** the user opens the SettingsModal
- **THEN** the Account section is displayed and highlighted as active in the left nav

#### Scenario: Switching sections

- **WHEN** the user clicks "App", "Advanced", or "Extensions" in the left nav
- **THEN** the corresponding section's content replaces the displayed content
- **AND** that nav item is highlighted as active

#### Scenario: Closing the modal

- **WHEN** the user activates the close control
- **THEN** the SettingsModal closes

## ADDED Requirements

### Requirement: Extensions section in Settings modal

The Settings modal SHALL include an **Extensions** section. The section SHALL display download links and install instructions for both Firefox and Chrome, using `EXTENSION_FIREFOX_URL` and `EXTENSION_CHROME_URL` from `src/lib/downloads.ts`.

The Firefox subsection SHALL:
- Provide a download link using `EXTENSION_FIREFOX_URL`
- Include fallback manual install instructions: `about:addons` → gear icon → "Install Add-on From File"

The Chrome subsection SHALL:
- Provide a download link using `EXTENSION_CHROME_URL`
- Include a brief developer mode note with the key steps: unzip, `chrome://extensions`, enable Developer mode, Load unpacked

#### Scenario: Extensions section renders with both browser subsections

- **WHEN** the user navigates to the Extensions section in the Settings modal
- **THEN** Firefox and Chrome subsections are displayed, each with a download link and install instructions

#### Scenario: Firefox download link uses the correct URL constant

- **WHEN** the Extensions section is displayed
- **THEN** the Firefox download link's `href` matches `EXTENSION_FIREFOX_URL`

#### Scenario: Chrome download link uses the correct URL constant

- **WHEN** the Extensions section is displayed
- **THEN** the Chrome download link's `href` matches `EXTENSION_CHROME_URL`
