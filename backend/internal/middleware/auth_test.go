package middleware

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/MicahParks/jwkset"
	"github.com/devi/bookleaf/internal/domain"
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockUserUsecase struct {
	user *domain.User
	err  error
}

func (m *mockUserUsecase) GetOrProvision(_ context.Context, _ string) (*domain.User, error) {
	return m.user, m.err
}

func (m *mockUserUsecase) GetByID(_ context.Context, _ string) (*domain.User, error) {
	return m.user, m.err
}

func makeTestStorage(t *testing.T) (*jwkset.MemoryJWKSet, *rsa.PrivateKey) {
	t.Helper()
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	storage := jwkset.NewMemoryStorage()
	jwk, err := jwkset.NewJWKFromKey(&privateKey.PublicKey, jwkset.JWKOptions{
		Metadata: jwkset.JWKMetadataOptions{KID: "test-kid"},
	})
	require.NoError(t, err)
	require.NoError(t, storage.KeyWrite(context.Background(), jwk))

	return storage, privateKey
}

func makeSignedToken(t *testing.T, key *rsa.PrivateKey, issuer, audience string) string {
	t.Helper()
	claims := jwt.RegisteredClaims{
		Subject:   "kp_abc123",
		Issuer:    issuer,
		Audience:  jwt.ClaimStrings{audience},
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = "test-kid"
	tokenString, err := token.SignedString(key)
	require.NoError(t, err)
	return tokenString
}

func TestAuthMiddleware_ValidToken_SetsUserIDOnContext(t *testing.T) {
	storage, privateKey := makeTestStorage(t)
	mockUC := &mockUserUsecase{user: &domain.User{ID: "kp_abc123"}}
	mw := newAuthMiddlewareWithStorage("https://example.kinde.com", "bookleaf-api", storage, mockUC)
	tokenString := makeSignedToken(t, privateKey, "https://example.kinde.com", "bookleaf-api")

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	req.Header.Set(echo.HeaderAuthorization, "Bearer "+tokenString)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := mw(func(c echo.Context) error {
		userID, ok := AuthenticatedUserIDFromContext(c)
		require.True(t, ok)
		assert.Equal(t, "kp_abc123", userID)
		return c.NoContent(http.StatusNoContent)
	})(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, rec.Code)
}

func TestAuthMiddleware_InvalidOrMissingToken_Returns401(t *testing.T) {
	storage := jwkset.NewMemoryStorage()
	mockUC := &mockUserUsecase{user: &domain.User{ID: "kp_abc123"}}
	mw := newAuthMiddlewareWithStorage("https://example.kinde.com", "bookleaf-api", storage, mockUC)

	tests := []struct {
		name       string
		authHeader string
	}{
		{"missing authorization header", ""},
		{"invalid bearer token", "Bearer not-a-valid-jwt"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/me", nil)
			if tt.authHeader != "" {
				req.Header.Set(echo.HeaderAuthorization, tt.authHeader)
			}
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			nextCalled := false
			err := mw(func(c echo.Context) error {
				nextCalled = true
				return c.NoContent(http.StatusNoContent)
			})(c)

			httpErr, ok := err.(*echo.HTTPError)
			require.True(t, ok)
			assert.Equal(t, http.StatusUnauthorized, httpErr.Code)
			assert.False(t, nextCalled)
		})
	}
}

func TestAuthMiddleware_ProvisioningFailure_Returns500(t *testing.T) {
	storage, privateKey := makeTestStorage(t)
	mockUC := &mockUserUsecase{err: assert.AnError}
	mw := newAuthMiddlewareWithStorage("https://example.kinde.com", "bookleaf-api", storage, mockUC)
	tokenString := makeSignedToken(t, privateKey, "https://example.kinde.com", "bookleaf-api")

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	req.Header.Set(echo.HeaderAuthorization, "Bearer "+tokenString)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	nextCalled := false
	err := mw(func(c echo.Context) error {
		nextCalled = true
		return c.NoContent(http.StatusNoContent)
	})(c)

	httpErr, ok := err.(*echo.HTTPError)
	require.True(t, ok)
	assert.Equal(t, http.StatusInternalServerError, httpErr.Code)
	assert.False(t, nextCalled)
}
