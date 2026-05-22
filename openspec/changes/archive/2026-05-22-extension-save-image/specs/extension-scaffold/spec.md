## MODIFIED Requirements

### Requirement: Manifest V3 configuration

The extension manifest SHALL declare:
- `manifest_version: 3`
- `name`: "Bookleaf"
- `version`: "0.1.0"
- `permissions`: `["storage", "identity", "contextMenus", "notifications"]`
- `action` with `default_popup` pointing to the popup entry
- `background` with `service_worker` pointing to the background script entry
- `host_permissions`: `["<all_urls>"]`
- `icons` with entries for sizes `48` and `128` pointing to PNG files in `public/icons/`

#### Scenario: Manifest contains required permissions

- **WHEN** the extension is loaded in Chrome
- **THEN** the manifest is parsed without errors
- **AND** `storage`, `identity`, `contextMenus`, and `notifications` permissions are granted

### Requirement: Environment variable configuration

The project SHALL read configuration from environment variables via Vite's `import.meta.env`:
- `VITE_KINDE_CLIENT_ID`: The Kinde application client ID for the extension
- `VITE_KINDE_ISSUER_URL`: The Kinde issuer base URL (e.g., `https://<domain>.kinde.com`)
- `VITE_KINDE_AUDIENCE`: The API audience value expected by the Bookleaf backend JWT middleware
- `VITE_API_BASE_URL`: The Bookleaf backend base URL

An `.env.example` file SHALL be provided listing all required variables.

#### Scenario: Missing env var is detectable at build time

- **WHEN** `VITE_KINDE_CLIENT_ID` is not set
- **THEN** the build emits a warning or error identifying the missing variable
