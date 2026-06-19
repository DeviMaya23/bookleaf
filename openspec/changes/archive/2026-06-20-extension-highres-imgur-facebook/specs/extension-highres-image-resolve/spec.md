## ADDED Requirements

### Requirement: Imgur thumbnail image rule

A rule SHALL match `i.imgur.com` URLs whose path is a single segment of the form `{id}{sizeLetter}.{ext}` where `{sizeLetter}` is a single lowercase letter preceded by an underscore (e.g. `_d`) and `{ext}` is one of `webp`, `jpg`, `jpeg`, `png`, or `gif`. The rule SHALL NOT match Imgur video delivery URLs (paths ending in `.mp4` or `.gifv`) or paths that do not fit the `{id}_{letter}.{ext}` shape (e.g. an id with no size suffix). Its transform SHALL strip the size-letter suffix and any query string, and SHALL request the bare id with a `.jpg` extension.

#### Scenario: Imgur webp thumbnail resolves to the bare id with a jpg extension

- **WHEN** `resolveHighResUrl` is called with `https://i.imgur.com/xCbCj7a_d.webp?maxwidth=520&shape=thumb&fidelity=high`
- **THEN** it returns `https://i.imgur.com/xCbCj7a.jpg`

#### Scenario: Imgur URL with no size suffix does not match

- **WHEN** `resolveHighResUrl` is called with `https://i.imgur.com/xCbCj7a.jpg`
- **THEN** it returns `null`

#### Scenario: Imgur video delivery URL does not match

- **WHEN** `resolveHighResUrl` is called with an `i.imgur.com` URL ending in `.mp4` or `.gifv`
- **THEN** it returns `null`

### Requirement: Facebook CDN ctp-param rule

A rule SHALL match `fbcdn.net` image URLs (any subdomain) that carry a `ctp` query parameter. Its transform SHALL remove only the `ctp` parameter, leaving all other query parameters (including `stp`, `cstp`, and the signature-bearing `_nc_*`/`oh`/`oe` parameters) unchanged.

#### Scenario: Facebook CDN URL with ctp resolves with ctp removed

- **WHEN** `resolveHighResUrl` is called with a `scontent-*.xx.fbcdn.net` image URL containing `ctp=s590x590` alongside `cstp=mx1631x1087` and signature params `oh`/`oe`
- **THEN** it returns the same URL with `ctp` removed and every other parameter (including `cstp`, `oh`, `oe`) unchanged

#### Scenario: Facebook CDN URL without ctp does not match

- **WHEN** `resolveHighResUrl` is called with an `fbcdn.net` image URL that has no `ctp` parameter
- **THEN** it returns `null`
