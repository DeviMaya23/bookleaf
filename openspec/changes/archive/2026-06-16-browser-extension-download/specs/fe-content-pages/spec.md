## ADDED Requirements

### Requirement: Extensions page

The app SHALL expose an `/extensions` route rendering a page built on `SimplePageLayout` with the title "Browser Extension". The page SHALL contain two sections — Firefox and Chrome — each with a download link sourced from `src/lib/downloads.ts` and install instructions appropriate for that browser.

The Firefox section SHALL:
- Provide a download link using `EXTENSION_FIREFOX_URL`
- Include fallback manual install instructions: open `about:addons`, click the gear icon, select "Install Add-on From File", and choose the downloaded `.xpi`

The Chrome section SHALL:
- Provide a download link using `EXTENSION_CHROME_URL`
- Include explicit step-by-step developer mode instructions: download and unzip the file, navigate to `chrome://extensions`, enable Developer mode, click "Load unpacked", and select the unzipped folder
- State clearly that Chrome Web Store approval is pending

The landing page footer SHALL include an "Extensions" link to `/extensions` alongside the existing About, Privacy, and AI Notes links.

#### Scenario: Extensions page renders with both browser sections

- **WHEN** the user navigates to `/extensions`
- **THEN** the page renders using `SimplePageLayout` with sections for Firefox and Chrome, each containing a download link and install instructions

#### Scenario: Firefox download link uses the correct URL constant

- **WHEN** the Extensions page is rendered
- **THEN** the Firefox download link's `href` matches `EXTENSION_FIREFOX_URL`

#### Scenario: Chrome download link uses the correct URL constant

- **WHEN** the Extensions page is rendered
- **THEN** the Chrome download link's `href` matches `EXTENSION_CHROME_URL`

#### Scenario: Extensions link appears in the landing page footer

- **WHEN** the user is on the landing page
- **THEN** an "Extensions" link to `/extensions` is present in the footer
