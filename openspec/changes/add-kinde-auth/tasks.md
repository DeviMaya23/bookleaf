## 1. Dependencies

- [ ] 1.1 Add `github.com/golang-jwt/jwt/v5` to `backend/go.mod`
- [ ] 1.2 Add a JWKS client library (e.g. `github.com/MicahParks/jwkset`) to `backend/go.mod`
- [ ] 1.3 Run `go mod tidy` in `backend/`

## 2. User Repository

- [ ] 2.1 Define `UserRepository` interface in `backend/internal/usecase/` with `GetOrCreate(ctx, id string) (*domain.User, error)`
- [ ] 2.2 Implement `UserRepository` in `backend/internal/repository/` using `INSERT ... ON CONFLICT DO NOTHING` followed by `SELECT`
- [ ] 2.3 Write unit test for repository: happy path (new user created) and error path (DB error returns error)

## 3. User Usecase

- [ ] 3.1 Implement `UserUsecase` in `backend/internal/usecase/` with `GetOrProvision(ctx, kindeID string) (*domain.User, error)` calling the repository
- [ ] 3.2 Write unit test for usecase: happy path (user returned) and error path (repository error propagates)

## 4. Auth Middleware

- [ ] 4.1 Create `backend/internal/middleware/auth.go` with typed context key constant
- [ ] 4.2 Implement JWKS client initialisation using `KINDE_ISSUER_URL` to fetch `/.well-known/jwks`
- [ ] 4.3 Implement JWT validation: verify signature, expiry, issuer, and audience (`KINDE_AUDIENCE`)
- [ ] 4.4 Implement user auto-provisioning call inside middleware using `UserUsecase`
- [ ] 4.5 Set authenticated user ID on Echo context using typed key
- [ ] 4.6 Write unit test for middleware: happy path (valid token sets user ID on context) and error path (missing/invalid token returns 401)

## 5. Me Handler

- [ ] 5.1 Create `backend/internal/handler/me.go` implementing `GET /me`
- [ ] 5.2 Handler reads user ID from Echo context, fetches user from usecase, returns `{ "id", "vision_enabled" }`
- [ ] 5.3 Write unit test for handler: happy path (returns 200 with correct body) and error path (usecase error returns 500)

## 6. Server Wiring

- [ ] 6.1 Add env var validation in `backend/cmd/server/main.go` — fail fast if `KINDE_ISSUER_URL` or `KINDE_AUDIENCE` missing
- [ ] 6.2 Initialise JWKS client and `UserUsecase` in `main.go`
- [ ] 6.3 Register protected route group with auth middleware
- [ ] 6.4 Register `GET /me` on the protected group
