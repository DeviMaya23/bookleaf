package middleware

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

const internalSecretHeader = "X-Bookleaf-Internal-Secret"

// NewInternalSecretMiddleware gates requests behind a shared secret header.
// Requests missing the header or supplying an incorrect value receive 401.
func NewInternalSecretMiddleware(secret string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if c.Request().Header.Get(internalSecretHeader) != secret {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			}
			return next(c)
		}
	}
}
