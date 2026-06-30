-include Makefile.local

.PHONY: tidy run test-cover-repository rebuild fe-install fe-dev fe-test ext-install ext-build ext-build-firefox ext-build-all ext-build-prod

tidy:
	@cd backend && go mod tidy

run:
	@cd backend && go run ./cmd/server

test-cover-repository:
	@cd backend && go test -covermode=atomic -coverprofile=internal/repository/coverage.out ./internal/repository/...

rebuild:
	@docker compose build --no-cache app && docker compose up -d app

fe-install:
	@cd frontend && npm install

fe-dev:
	@cd frontend && npm run dev

fe-test:
	@cd frontend && npm run test

be-dev:
	@cd backend && go run ./cmd/server

ext-install:
	@cd extensions && npm install

ext-build-chrome:
	@cd extensions && npm run build

ext-build-firefox:
	@cd extensions && npm run build:firefox

ext-build:
	@cd extensions && npm run build:all

ext-build-prod:
	@cd extensions && npm run build:chrome:prod && npm run build:firefox:prod
update-clearurls:
	@curl -fsSL https://rules2.clearurls.xyz/data.min.json -o extensions/vendor/clearurls-data.min.json
	@jq '{providers: {google: .providers.google, duckduckgo: .providers.duckduckgo, twitter: .providers.twitter, instagram: .providers.instagram, facebook: .providers.facebook, reddit: .providers.reddit}}' extensions/vendor/clearurls-data.min.json > extensions/src/lib/clearUrlsProviders.json
