package middleware

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/MicahParks/jwkset"
	"github.com/devi/bookleaf/internal/domain"
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

type mockUserUsecase struct {
	user     *domain.User
	err      error
	calledID string
}

func (m *mockUserUsecase) GetOrProvision(_ context.Context, kindeID string) (*domain.User, error) {
	m.calledID = kindeID
	return m.user, m.err
}

func (m *mockUserUsecase) GetByID(_ context.Context, _ string) (*domain.User, error) {
	return m.user, m.err
}

func TestAuthMiddleware_Success(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate rsa key: %v", err)
	}

	jwksStorage := jwkset.NewMemoryStorage()
	jwk, err := jwkset.NewJWKFromKey(
		&privateKey.PublicKey,
		jwkset.JWKOptions{
			Metadata: jwkset.JWKMetadataOptions{
				KID: "test-kid",
			},
		},
	)
	if err != nil {
		t.Fatalf("create jwk: %v", err)
	}
	if err := jwksStorage.KeyWrite(context.Background(), jwk); err != nil {
		t.Fatalf("write jwk: %v", err)
	}

	mockUC := &mockUserUsecase{
		user: &domain.User{ID: "kp_abc123", VisionEnabled: false},
	}

	mw := newAuthMiddlewareWithStorage(
		"https://example.kinde.com",
		"bookleaf-api",
		jwksStorage,
		mockUC,
	)

	claims := jwt.RegisteredClaims{
		Subject:   "kp_abc123",
		Issuer:    "https://example.kinde.com",
		Audience:  jwt.ClaimStrings{"bookleaf-api"},
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = "test-kid"
	tokenString, err := token.SignedString(privateKey)
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	req.Header.Set(echo.HeaderAuthorization, "Bearer "+tokenString)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err = mw(func(c echo.Context) error {
		userID, ok := AuthenticatedUserIDFromContext(c)
		if !ok {
			t.Fatal("expected user ID in context")
		}
		if userID != "kp_abc123" {
			t.Fatalf("expected user ID kp_abc123, got %s", userID)
		}
		return c.NoContent(http.StatusNoContent)
	})(c)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, rec.Code)
	}
	if mockUC.calledID != "kp_abc123" {
		t.Fatalf("expected GetOrProvision called with kp_abc123, got %s", mockUC.calledID)
	}
}

func TestAuthMiddleware_UnauthorizedOnMissingOrInvalidToken(t *testing.T) {
	jwksStorage := jwkset.NewMemoryStorage()
	mockUC := &mockUserUsecase{
		user: &domain.User{ID: "kp_abc123", VisionEnabled: false},
	}

	mw := newAuthMiddlewareWithStorage(
		"https://example.kinde.com",
		"bookleaf-api",
		jwksStorage,
		mockUC,
	)

	tests := []struct {
		name         string
		authHeader   string
		expectedCode int
	}{
		{
			name:         "missing authorization header",
			authHeader:   "",
			expectedCode: http.StatusUnauthorized,
		},
		{
			name:         "invalid bearer token",
			authHeader:   "Bearer not-a-valid-jwt",
			expectedCode: http.StatusUnauthorized,
		},
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
			if err == nil {
				t.Fatal("expected an unauthorized error, got nil")
			}

			httpErr, ok := err.(*echo.HTTPError)
			if !ok {
				t.Fatalf("expected *echo.HTTPError, got %T", err)
			}
			if httpErr.Code != tt.expectedCode {
				t.Fatalf("expected status %d, got %d", tt.expectedCode, httpErr.Code)
			}
			if nextCalled {
				t.Fatal("expected next handler not to be called")
			}
		})
	}
}

func TestAuthMiddleware_ProvisioningFailureReturnsInternalServerError(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate rsa key: %v", err)
	}

	jwksStorage := jwkset.NewMemoryStorage()
	jwk, err := jwkset.NewJWKFromKey(
		&privateKey.PublicKey,
		jwkset.JWKOptions{
			Metadata: jwkset.JWKMetadataOptions{
				KID: "test-kid",
			},
		},
	)
	if err != nil {
		t.Fatalf("create jwk: %v", err)
	}
	if err := jwksStorage.KeyWrite(context.Background(), jwk); err != nil {
		t.Fatalf("write jwk: %v", err)
	}

	mockUC := &mockUserUsecase{
		err: errors.New("db error"),
	}

	mw := newAuthMiddlewareWithStorage(
		"https://example.kinde.com",
		"bookleaf-api",
		jwksStorage,
		mockUC,
	)

	claims := jwt.RegisteredClaims{
		Subject:   "kp_abc123",
		Issuer:    "https://example.kinde.com",
		Audience:  jwt.ClaimStrings{"bookleaf-api"},
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = "test-kid"
	tokenString, err := token.SignedString(privateKey)
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	req.Header.Set(echo.HeaderAuthorization, "Bearer "+tokenString)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	nextCalled := false
	err = mw(func(c echo.Context) error {
		nextCalled = true
		return c.NoContent(http.StatusNoContent)
	})(c)
	if err == nil {
		t.Fatal("expected an internal server error, got nil")
	}

	httpErr, ok := err.(*echo.HTTPError)
	if !ok {
		t.Fatalf("expected *echo.HTTPError, got %T", err)
	}
	if httpErr.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, httpErr.Code)
	}
	if nextCalled {
		t.Fatal("expected next handler not to be called")
	}
}
