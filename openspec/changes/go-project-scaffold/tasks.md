## 1. Initialize Go module

- [ ] 1.1 Run `go mod init github.com/devi/bookleaf`
- [ ] 1.2 Add Echo as a dependency (`go get github.com/labstack/echo/v4`)

## 2. Create directory structure

- [ ] 2.1 Create `cmd/server/`
- [ ] 2.2 Create `internal/domain/`, `internal/usecase/`, `internal/repository/`, `internal/handler/` each with a `.gitkeep`

## 3. Write entry point

- [ ] 3.1 Write `cmd/server/main.go` — bootstrap Echo, read `PORT` from env, register `GET /health`, start server

## 4. Verify

- [ ] 4.1 `go run ./cmd/server` starts without errors
- [ ] 4.2 `curl localhost:8080/health` returns `200 OK`
