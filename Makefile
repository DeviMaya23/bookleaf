-include Makefile.local

.PHONY: tidy run test-cover-repository compose-rebuild-app

tidy:
	@cd backend && go mod tidy

run:
	@cd backend && go run ./cmd/server

test-cover-repository:
	@cd backend && go test -covermode=atomic -coverprofile=internal/repository/coverage.out ./internal/repository/...

compose-rebuild-app:
	@docker compose build --no-cache app
