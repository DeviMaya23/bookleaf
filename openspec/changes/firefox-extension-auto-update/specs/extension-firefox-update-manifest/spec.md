## ADDED Requirements

### Requirement: Update manifest generation produces a Mozilla-format `updates.json` matching the signed `.xpi`

After `scripts/sign-firefox.sh` produces a signed Firefox production `.xpi` via `web-ext sign`, it SHALL generate an `updates.json` file containing a single update entry for the production gecko ID (`bookleaf@evimay.me`) with: the `version` read from `extensions/manifest.json`, an `update_link` pointing at the fixed Firefox `.xpi` URL, and an `update_hash` of the form `sha256:<hex digest>` computed from the `.xpi` file that was just signed (located by freshest modification time in `web-ext-artifacts/`). Each run SHALL overwrite any previous `updates.json` with this single entry — no historical versions are retained.

Hashing SHALL prefer `sha256sum` if available on `PATH`, falling back to `shasum -a 256` otherwise.

#### Scenario: Update manifest is generated after a successful sign

- **WHEN** `npm run sign:firefox` completes signing a `.xpi`
- **THEN** `web-ext-artifacts/updates.json` exists
- **AND** its `addons["bookleaf@evimay.me"].updates` array contains exactly one entry
- **AND** that entry's `version` matches the `version` field in `extensions/manifest.json`
- **AND** that entry's `update_hash` is `sha256:` followed by the sha256 digest of the signed `.xpi` file
- **AND** that entry's `update_link` is the fixed Firefox `.xpi` download URL

#### Scenario: Re-running the release script overwrites the previous update manifest

- **GIVEN** `web-ext-artifacts/updates.json` already exists from a prior release with an older version
- **WHEN** `npm run sign:firefox` is run again after the version in `extensions/manifest.json` was bumped
- **THEN** `web-ext-artifacts/updates.json` contains exactly one `updates` entry reflecting the new version and the newly signed `.xpi`'s hash
- **AND** no entry for the prior version remains in the file

#### Scenario: Hashing falls back to shasum when sha256sum is unavailable

- **WHEN** the release script computes the `.xpi` hash on a machine without `sha256sum` on `PATH`
- **THEN** the script uses `shasum -a 256` instead
- **AND** the resulting `update_hash` value is unaffected by which tool computed it
