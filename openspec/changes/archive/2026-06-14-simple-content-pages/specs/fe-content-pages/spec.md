## ADDED Requirements

### Requirement: SimplePageLayout shared layout

The app SHALL provide a `SimplePageLayout` component for plain text-content pages. It SHALL render a nav bar containing the "Bookleaf" wordmark as a link to `/`, followed by a main content area containing a page title and body content, centered in a max-width column. The page SHALL scroll normally (no fixed-viewport constraint).

#### Scenario: Layout renders nav and title

- **WHEN** a page built on `SimplePageLayout` is rendered with a title and body content
- **THEN** a nav bar with the "Bookleaf" wordmark is shown, followed by the title and body content in a centered column

#### Scenario: Wordmark links back to the landing page

- **WHEN** the user clicks the "Bookleaf" wordmark in the nav bar
- **THEN** the browser navigates to `/`

#### Scenario: Page content can exceed the viewport height

- **WHEN** the page's body content is taller than the viewport
- **THEN** the page scrolls normally to reveal the remaining content

### Requirement: About page

The app SHALL expose an `/about` route rendering a page built on `SimplePageLayout`, with section headings describing what the page will cover and placeholder body text for each section.

#### Scenario: About page renders

- **WHEN** the user navigates to `/about`
- **THEN** the page renders using `SimplePageLayout` with a title and one or more section headings, each followed by placeholder body text

### Requirement: Privacy Policy page

The app SHALL expose a `/privacy` route rendering a page built on `SimplePageLayout`, with section headings covering what data is collected, how it's used, and user choices, each followed by placeholder body text. The page SHALL include a link to the AI Notes page (`/ai-notes`).

#### Scenario: Privacy Policy page renders

- **WHEN** the user navigates to `/privacy`
- **THEN** the page renders using `SimplePageLayout` with section headings covering data collection, usage, and user choices, each followed by placeholder body text

#### Scenario: Privacy Policy links to AI Notes

- **WHEN** the user is on the Privacy Policy page
- **THEN** a link to `/ai-notes` is present

### Requirement: AI Notes page

The app SHALL expose an `/ai-notes` route rendering a page built on `SimplePageLayout`, with section headings covering how AI features work, what is sent to Google Vision API, and how to opt out, each followed by placeholder body text. The page SHALL include a link to the Privacy Policy page (`/privacy`).

#### Scenario: AI Notes page renders

- **WHEN** the user navigates to `/ai-notes`
- **THEN** the page renders using `SimplePageLayout` with section headings covering how AI features work, what is sent to Google Vision API, and how to opt out, each followed by placeholder body text

#### Scenario: AI Notes links to Privacy Policy

- **WHEN** the user is on the AI Notes page
- **THEN** a link to `/privacy` is present
