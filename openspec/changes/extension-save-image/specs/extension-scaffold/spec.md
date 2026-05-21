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

#### Scenario: Manifest contains required permissions

- **WHEN** the extension is loaded in Chrome
- **THEN** the manifest is parsed without errors
- **AND** `storage`, `identity`, `contextMenus`, and `notifications` permissions are granted
