-include Makefile.local

.PHONY: tidy run test-cover-repository rebuild fe-install fe-dev ext-install ext-build ext-build-firefox ext-build-all

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