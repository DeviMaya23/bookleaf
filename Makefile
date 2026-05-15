-include Makefile.local

.PHONY: tidy run test-cover-repository rebuild fe-install fe-dev

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
