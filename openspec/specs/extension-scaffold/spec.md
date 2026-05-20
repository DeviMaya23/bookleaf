# Spec: Extension Scaffold

## Purpose

Defines the scaffolding requirements for the Bookleaf browser extension project, including project structure, build configuration, manifest, TypeScript setup, and environment variable handling.

## Requirements

### Requirement: Extension project structure

The system SHALL provide a browser extension project under `/extensions` at the repository root. The project SHALL use Vite as the build tool, TypeScript as the language, and `vite-plugin-web-extension` to handle multi-entry MV3 builds. The project SHALL include `webextension-polyfill` and its TypeScript types to enable cross-browser compatibility.

#### Scenario: Project builds successfully for Chrome

- **WHEN** `npm run build` is run in `/extensions`
- **THEN** a production build is produced in `/extensions/dist/` with a valid `manifest.json`
- **AND** the manifest declares `manifest_version: 3`

#### Scenario: Project builds successfully for Firefox

- **WHEN** `npm run build:firefox` is run in `/extensions`
- **THEN** a build is produced targeting Firefox MV3
- **AND** the manifest is valid for Firefox

### Requirement: Manifest V3 configuration

The extension manifest SHALL declare:
- `manifest_version: 3`
- `name`: "Bookleaf"
- `version`: "0.1.0"
- `permissions`: `["storage", "identity"]`
- `action` with `default_popup` pointing to the popup entry
- `background` with `service_worker` pointing to the background script entry
- `host_permissions`: `["<all_urls>"]` (required for future image-saving feature)

#### Scenario: Manifest contains required permissions

- **WHEN** the extension is loaded in Chrome
- **THEN** the manifest is parsed without errors
- **AND** `storage` and `identity` permissions are granted

### Requirement: TypeScript configuration

The project SHALL include a `tsconfig.json` configured for browser extension development:
- `target`: ES2020 or higher (required for service workers)
- `lib` includes `"DOM"` and `"WebWorker"` (popup needs DOM; service worker needs WebWorker)
- Strict mode enabled (`"strict": true`)

#### Scenario: TypeScript compiles without errors

- **WHEN** `tsc --noEmit` is run
- **THEN** the project compiles without type errors

### Requirement: Environment variable configuration

The project SHALL read configuration from environment variables via Vite's `import.meta.env`:
- `VITE_KINDE_CLIENT_ID`: The Kinde application client ID for the extension
- `VITE_KINDE_ISSUER_URL`: The Kinde issuer base URL (e.g., `https://<domain>.kinde.com`)
- `VITE_API_BASE_URL`: The Bookleaf backend base URL

An `.env.example` file SHALL be provided listing all required variables.

#### Scenario: Missing env var is detectable at build time

- **WHEN** `VITE_KINDE_CLIENT_ID` is not set
- **THEN** the build emits a warning or error identifying the missing variable
