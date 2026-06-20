## 1. Card resolution registration

- [x] 1.1 Add `instagramRule: CardDomResolveRule` to `extensions/src/lib/cardDomResolveRules.ts`, matching `url.hostname === "instagram.com"` or `url.hostname.endsWith(".instagram.com")`.
- [x] 1.2 Add `instagramRule` to the exported `rules` array.
- [x] 1.3 Add `"*://*.instagram.com/p/*"` to the exported `linkOnlyCardUrlPatterns` array.

## 2. Unit tests

- [x] 2.1 In `extensions/src/lib/cardDomResolveRules.test.ts`, add to the `shouldResolveCardDom` describe block: matches `https://www.instagram.com/p/Cxxxxx/`, matches an Instagram locale/subdomain variant, and confirm the existing "does not match an unregistered site" case still passes unaffected.

## 3. Verification

- [x] 3.1 Run `npm run build` in `extensions/` and fix any errors.
- [x] 3.2 Run `npm run lint` in `extensions/` and fix any errors. (No `lint` script exists in `extensions/`; ran `npm run type-check` instead — passed with no errors.)
- [x] 3.3 Run the extension test suite (`npm test` / vitest) in `extensions/` and confirm all tests pass.
