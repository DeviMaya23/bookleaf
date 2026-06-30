# Spec: Extension Source URL Cleaning

## Purpose

Defines how the extension strips tracking query parameters from `source_url` values before sending them to the backend, using a vendored subset of ClearURLs provider rules.

## Requirements

### Requirement: Vendored ClearURLs provider subset

The extension repository SHALL maintain the following ClearURLs-related files:

1. `extensions/vendor/clearurls-data.min.json` — a full snapshot of the ClearURLs rule database, fetched from `https://rules2.clearurls.xyz/data.min.json` by `make update-clearurls`. This file is gitignored and SHALL NOT be imported by any extension source file.

2. `extensions/src/lib/clearUrlsProviders.json` — a 6-provider subset extracted from the full snapshot, committed to the repository, imported by `urlCleaner.ts`, and included in the extension bundle. The file SHALL contain a top-level `_comment` field explaining its source and update command. The providers are: `google`, `duckduckgo`, `twitter`, `instagram`, `facebook`, `reddit`. Pinterest and Imgur are excluded as ClearURLs has no rules for them.

Running `make update-clearurls` re-fetches the full snapshot and re-extracts the subset, after which only `clearUrlsProviders.json` is committed.

#### Scenario: Bundled subset contains only the target providers

- **WHEN** `clearUrlsProviders.json` is loaded
- **THEN** its `providers` object contains keys for exactly: `google`, `duckduckgo`, `twitter`, `instagram`, `facebook`, `reddit`

### Requirement: ClearURLs rule application

The `cleanUrl(rawUrl: string): string` function in `extensions/src/lib/urlCleaner.ts` SHALL apply ClearURLs provider rules to strip tracking query params from a URL. The function SHALL:

1. Parse `rawUrl` into a `URL` object. If parsing throws, return `rawUrl` unchanged.
2. Iterate all providers in the vendored subset. For each provider:
   a. Test the provider's `urlPattern` regex against the full URL string. If it does not match, skip this provider.
   b. Test each of the provider's `exceptions` regexes against the full URL string. If any exception matches, skip this provider entirely.
   c. If `completeProvider` is `true`, delete all query params and return the URL immediately.
   d. Otherwise, for each query param name in the URL, delete it if its name fully matches any regex in the provider's `rules` or `referralMarketing` arrays.
3. After iterating all providers, return the cleaned URL as a string.

Rule strings from the ClearURLs data are treated as regex patterns anchored with `^` and `$` when tested against param names (e.g. `"utm_.*"` matches any param name starting with `utm_`). Regex compilation from the JSON strings SHALL occur once at module load time, not on each `cleanUrl()` call.

#### Scenario: Known tracking param is stripped

- **WHEN** `cleanUrl` receives a URL whose host matches a provider and the URL contains a param name matching one of that provider's `rules` entries
- **THEN** the matching param is absent from the returned URL
- **AND** non-matching params are preserved

#### Scenario: `referralMarketing` params are also stripped

- **WHEN** `cleanUrl` receives a URL whose host matches a provider and the URL contains a param name matching one of that provider's `referralMarketing` entries
- **THEN** the matching param is absent from the returned URL

#### Scenario: Provider exception vetoes rule application

- **WHEN** `cleanUrl` receives a URL that matches a provider's `urlPattern` but also matches one of that provider's `exceptions`
- **THEN** none of that provider's rules are applied and the URL's params are left unchanged by that provider

#### Scenario: `completeProvider` strips all query params

- **WHEN** `cleanUrl` receives a URL matching a provider whose `completeProvider` is `true`
- **THEN** the returned URL has no query params

#### Scenario: Multiple providers match and both are applied

- **WHEN** `cleanUrl` receives a URL that matches more than one provider in the subset
- **THEN** rules from all matching providers are applied to the URL

#### Scenario: URL from an unrecognised host is returned unchanged

- **WHEN** `cleanUrl` receives a URL whose host does not match any provider's `urlPattern`
- **THEN** the returned string is identical to the input

#### Scenario: Malformed input is returned unchanged

- **WHEN** `cleanUrl` receives a string that cannot be parsed as a URL
- **THEN** the returned string is identical to the input
