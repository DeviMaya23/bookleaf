# Spec: Extension High-Resolution Image Resolve

## Purpose

Defines the site-specific URL rewrite rule table and resolver used to find a high-resolution candidate URL for a known site's thumbnail image URL, and the validation applied to that candidate before it is treated as usable, as part of the "Save to Bookleaf" extension save flow.

## Requirements

### Requirement: High-resolution URL rule table

The extension SHALL maintain a table of site-specific high-resolution URL rules, each consisting of an `id`, a `matches(url: URL): boolean` predicate, and a synchronous `transform(url: URL): string | null` function. A `resolveHighResUrl(srcUrl: string): string | null` function SHALL iterate the table in order and return the first matching rule's transform result, or `null` if no rule matches.

#### Scenario: No rule matches an unrecognized site

- **WHEN** `resolveHighResUrl` is called with a URL from a site with no registered rule
- **THEN** it returns `null`

### Requirement: Twitter/X media image rule

A rule SHALL match `pbs.twimg.com` URLs that contain a `/media/` path segment, carry a `format` query param of `jpg`, `png`, or `webp`, and carry a `name` query param that is present and not already `orig`. The rule SHALL NOT match profile or banner image URLs (those under `/profile_images/` or `/profile_banners/` instead of `/media/`). Its transform SHALL set the `name` query param to `orig`, leaving the rest of the URL unchanged.

#### Scenario: Twitter media image with name=small resolves to name=orig

- **WHEN** `resolveHighResUrl` is called with `https://pbs.twimg.com/media/XXXXX?format=jpg&name=small`
- **THEN** it returns `https://pbs.twimg.com/media/XXXXX?format=jpg&name=orig`

#### Scenario: Twitter media image with a numeric size name resolves to name=orig

- **WHEN** `resolveHighResUrl` is called with `https://pbs.twimg.com/media/XXXXX?format=jpg&name=360x360`
- **THEN** it returns `https://pbs.twimg.com/media/XXXXX?format=jpg&name=orig`

#### Scenario: Twitter profile image does not match the media rule

- **WHEN** `resolveHighResUrl` is called with `https://pbs.twimg.com/profile_images/XXXXX/avatar_normal.jpg`
- **THEN** it returns `null`

### Requirement: Pinterest size-segment rule

A rule SHALL match `i.pinimg.com` URLs whose path contains a size-segment component (e.g. `236x`, `474x`, `564x`, `736x`). Its transform SHALL replace that size segment with `originals`.

#### Scenario: Pinterest sized image resolves to originals

- **WHEN** `resolveHighResUrl` is called with `https://i.pinimg.com/736x/aa/bb/cc/aabbccdd.jpg`
- **THEN** it returns `https://i.pinimg.com/originals/aa/bb/cc/aabbccdd.jpg`

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

### Requirement: High-resolution candidate validation

When a candidate high-resolution URL is fetched, the extension SHALL validate the response before treating it as usable: the response status SHALL be `ok`, the `Content-Type` SHALL be one of `image/jpeg`, `image/png`, or `image/webp`, and (when `createImageBitmap` is available) the decoded pixel dimensions SHALL be at least 100×100. A candidate that fails any check SHALL be treated as invalid. When `createImageBitmap` is unavailable, the dimension check SHALL be skipped and only the status and content-type checks apply.

#### Scenario: Valid high-resolution response passes validation

- **WHEN** a candidate URL's response is `ok`, has `Content-Type: image/jpeg`, and decodes to dimensions of at least 100×100
- **THEN** the candidate is treated as valid and its blob is used for the save

#### Scenario: Non-OK response fails validation

- **WHEN** a candidate URL's response is not `ok`
- **THEN** the candidate is treated as invalid

#### Scenario: Disallowed content-type fails validation

- **WHEN** a candidate URL's response has a `Content-Type` outside `image/jpeg`, `image/png`, `image/webp`
- **THEN** the candidate is treated as invalid

#### Scenario: Undersized decoded dimensions fail validation

- **WHEN** a candidate URL's response decodes to pixel dimensions smaller than 100×100 in either dimension
- **THEN** the candidate is treated as invalid

#### Scenario: Dimension check is skipped when createImageBitmap is unavailable

- **WHEN** a candidate URL's response is `ok` with an allowed `Content-Type`, and `createImageBitmap` is not available in the runtime
- **THEN** the candidate is treated as valid based on status and content-type alone
