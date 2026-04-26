## 1. Initialize Go module

- [x] 1.1 Run `go mod init github.com/devi/bookleaf`
- [x] 1.2 Add Echo as a dependency (`go get github.com/labstack/echo/v4`)

## 2. Create directory structure

- [x] 2.1 Create `cmd/server/`
- [x] 2.2 Create `internal/domain/`, `internal/usecase/`, `internal/repository/`, `internal/handler/` each with a `.gitkeep`

## 3. Write entry point

- [x] 3.1 Write `cmd/server/main.go` — bootstrap Echo, read `PORT` from env, register `GET /health`, start server

## 4. Verify

- [x] 4.1 `go run ./cmd/server` starts without errors
- [x] 4.2 `curl localhost:8080/health` returns `200 OK`
