## ADDED Requirements

### Requirement: Vision Config Sub-Struct

The `Config` struct SHALL include a `Vision VisionConfig` field. `VisionConfig` SHALL have:

- `APIKey string` — loaded from `GOOGLE_VISION_API_KEY`; **optional** (empty string if unset)

`config.Load()` SHALL NOT return an error if `GOOGLE_VISION_API_KEY` is absent. When `APIKey` is empty, the application starts normally and Vision features are skipped at runtime.

#### Scenario: GOOGLE_VISION_API_KEY is set

- **WHEN** `GOOGLE_VISION_API_KEY=abc123` is present in the environment
- **THEN** `cfg.Vision.APIKey` is `"abc123"`

#### Scenario: GOOGLE_VISION_API_KEY is absent

- **WHEN** `GOOGLE_VISION_API_KEY` is not set
- **THEN** `config.Load()` returns a non-nil `*Config` with `cfg.Vision.APIKey` equal to `""`
- **AND** `config.Load()` returns a nil error
