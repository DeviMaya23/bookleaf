package middleware

import (
	"net/http"

	"github.com/devi/bookleaf/internal/platform/config"
	"github.com/labstack/echo/v4"
)

const maintenanceBypassHeader = "X-Bookleaf-Bypass"
const maintenanceHeader = "X-Bookleaf-Maintenance"

// NewMaintenanceMiddleware gates requests behind cfg.Enabled, unless the
// request's X-Bookleaf-Bypass header matches a non-empty cfg.BypassToken.
func NewMaintenanceMiddleware(cfg config.MaintenanceConfig) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			bypass := c.Request().Header.Get(maintenanceBypassHeader)
			if cfg.BypassToken != "" && bypass == cfg.BypassToken {
				return next(c)
			}

			if cfg.Enabled {
				c.Response().Header().Set(maintenanceHeader, "true")
				return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "maintenance"})
			}

			return next(c)
		}
	}
}
