package middleware

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/MicahParks/jwkset"
	"github.com/devi/bookleaf/internal/observability"
	"github.com/devi/bookleaf/internal/usecase"
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type ContextKey string

const AuthenticatedUserIDContextKey ContextKey = "authenticatedUserID"

type authMiddleware struct {
	issuerURL   string
	audience    string
	jwksClient  jwkset.Storage
	userUsecase usecase.UserUsecase
	logger      *zap.Logger
}

func NewAuthMiddleware(
	issuerURL string,
	audience string,
	userUsecase usecase.UserUsecase,
	logger *zap.Logger,
) (echo.MiddlewareFunc, error) {
	jwksURL := strings.TrimRight(issuerURL, "/") + "/.well-known/jwks"

	jwksClient, err := jwkset.NewDefaultHTTPClient([]string{jwksURL})
	if err != nil {
		return nil, fmt.Errorf("initialise jwks client: %w", err)
	}

	return newAuthMiddlewareWithStorage(issuerURL, audience, jwksClient, userUsecase, logger), nil
}

func newAuthMiddlewareWithStorage(
	issuerURL string,
	audience string,
	jwksClient jwkset.Storage,
	userUsecase usecase.UserUsecase,
	logger *zap.Logger,
) echo.MiddlewareFunc {
	if logger == nil {
		logger = zap.NewNop()
	}

	m := &authMiddleware{
		issuerURL:   issuerURL,
		audience:    audience,
		jwksClient:  jwksClient,
		userUsecase: userUsecase,
		logger:      logger,
	}

	return m.handle
}

func AuthenticatedUserIDFromContext(c echo.Context) (string, bool) {
	userID, ok := c.Get(string(AuthenticatedUserIDContextKey)).(string)
	return userID, ok
}

func (m *authMiddleware) handle(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		tokenString, err := extractBearerToken(c.Request().Header.Get(echo.HeaderAuthorization))
		if err != nil {
			observability.LoggerFromContext(c.Request().Context(), m.logger).Warn(
				"auth token rejected",
				zap.String("event", "auth.token_rejected"),
				zap.String("reason", "missing_header"),
			)
			return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
		}

		claims := &jwt.RegisteredClaims{}
		token, err := jwt.ParseWithClaims(
			tokenString,
			claims,
			m.lookupKey,
			jwt.WithValidMethods([]string{jwt.SigningMethodRS256.Alg()}),
			jwt.WithIssuer(m.issuerURL),
			jwt.WithAudience(m.audience),
		)
		if err != nil || !token.Valid {
			observability.LoggerFromContext(c.Request().Context(), m.logger).Warn(
				"auth token rejected",
				zap.String("event", "auth.token_rejected"),
				zap.String("reason", "invalid_token"),
			)
			return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
		}

		if claims.Subject == "" {
			observability.LoggerFromContext(c.Request().Context(), m.logger).Warn(
				"auth token rejected",
				zap.String("event", "auth.token_rejected"),
				zap.String("reason", "missing_subject"),
			)
			return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
		}

		_, err = m.userUsecase.GetOrProvision(c.Request().Context(), claims.Subject)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to provision user")
		}

		c.Set(string(AuthenticatedUserIDContextKey), claims.Subject)
		return next(c)
	}
}

func (m *authMiddleware) lookupKey(token *jwt.Token) (any, error) {
	kidRaw, ok := token.Header["kid"]
	if !ok {
		return nil, errors.New("token missing kid header")
	}

	kid, ok := kidRaw.(string)
	if !ok || kid == "" {
		return nil, errors.New("token has invalid kid header")
	}

	jwk, err := m.jwksClient.KeyRead(context.Background(), kid)
	if err != nil {
		return nil, fmt.Errorf("read key from jwks: %w", err)
	}

	return jwk.Key(), nil
}

func extractBearerToken(authorizationHeader string) (string, error) {
	parts := strings.SplitN(strings.TrimSpace(authorizationHeader), " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || strings.TrimSpace(parts[1]) == "" {
		return "", errors.New("invalid authorization header")
	}

	return strings.TrimSpace(parts[1]), nil
}
