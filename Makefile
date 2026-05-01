-include Makefile.local

.PHONY: tidy run

tidy:
	@cd backend && go mod tidy

run:
	@cd backend && go run ./cmd/server
