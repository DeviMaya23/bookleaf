## ADDED Requirements

### Requirement: Public landing page at root

The app SHALL render a public landing page at `/` for unauthenticated visitors, consisting of a nav bar, a hero section, a features section, and a footer, laid out within a single non-scrollable viewport (`100dvh`, no page scroll).

#### Scenario: Unauthenticated user sees the landing page

- **WHEN** an unauthenticated user navigates to `/`
- **THEN** the landing page renders with a nav bar, hero section, features section, and footer

#### Scenario: Landing page fits within the viewport without scrolling

- **WHEN** the landing page is rendered at a typical desktop viewport size
- **THEN** the nav, hero, features, and footer are all visible without the page scrolling

### Requirement: Nav bar

The nav bar SHALL display the "Bookleaf" wordmark and a "Sign in" button. Clicking "Sign in" SHALL call `login()` from `useKindeAuth()`, redirecting the user to Kinde's hosted login page.

#### Scenario: Nav bar shows wordmark and sign-in button

- **WHEN** the landing page is rendered
- **THEN** the nav bar displays the "Bookleaf" wordmark and a "Sign in" button

#### Scenario: Sign-in button initiates Kinde login

- **WHEN** the user clicks the "Sign in" button in the nav bar
- **THEN** the browser redirects to Kinde's hosted login flow

### Requirement: Hero section

The hero section SHALL display an overline, a headline, a tagline, a "Get started" button, and a screenshot of the app. The screenshot SHALL be rendered at a 4:3 aspect ratio using `object-fit: cover`. Clicking "Get started" SHALL call `login()` from `useKindeAuth()`, redirecting the user to Kinde's hosted login page.

#### Scenario: Hero section displays copy and screenshot

- **WHEN** the landing page is rendered
- **THEN** the hero section displays an overline, headline, tagline, "Get started" button, and an app screenshot

#### Scenario: Hero screenshot renders at a 4:3 aspect ratio

- **WHEN** the landing page is rendered
- **THEN** the app screenshot occupies a 4:3 aspect-ratio box with its image content cropped via `object-fit: cover`

#### Scenario: Get-started button initiates Kinde login

- **WHEN** the user clicks the "Get started" button in the hero section
- **THEN** the browser redirects to Kinde's hosted login flow

### Requirement: Features section

The features section SHALL display exactly three feature items, each with a number, a title, and a short description, arranged in a three-column grid.

#### Scenario: Three feature items are displayed

- **WHEN** the landing page is rendered
- **THEN** the features section displays three items, each showing a number, title, and description

### Requirement: Authenticated users are redirected away from the landing page

The landing page SHALL NOT be accessible to already-authenticated users — they SHALL be redirected to `/app`.

#### Scenario: Authenticated user is redirected to the app

- **WHEN** an already-authenticated user navigates to `/`
- **THEN** they are redirected to `/app`

### Requirement: Error message display

The landing page SHALL display an error message when one is passed via React Router location state (e.g. after a failed callback).

#### Scenario: Error message is shown when passed via location state

- **WHEN** the user is redirected to `/` with an error message in React Router location state
- **THEN** the error message is displayed on the landing page
