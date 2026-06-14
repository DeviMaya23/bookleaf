package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/devi/bookleaf/internal/platform/config"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestInitEcho_CORSAllowsBypassHeader(t *testing.T) {
	cfg := &config.Config{CORSAllowedOrigins: []string{"http://localhost:5173"}}
	e := initEcho(cfg)

	req := httptest.NewRequest(http.MethodOptions, "/me", nil)
	req.Header.Set(echo.HeaderOrigin, "http://localhost:5173")
	req.Header.Set("Access-Control-Request-Method", http.MethodGet)
	req.Header.Set("Access-Control-Request-Headers", "X-Bookleaf-Bypass")
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.Contains(t, rec.Header().Get("Access-Control-Allow-Headers"), "X-Bookleaf-Bypass")
	assert.Contains(t, rec.Header().Get("Access-Control-Allow-Headers"), echo.HeaderAuthorization)
	assert.Contains(t, rec.Header().Get("Access-Control-Allow-Headers"), echo.HeaderContentType)
}

func TestInitEcho_CORSExposesMaintenanceHeader(t *testing.T) {
	cfg := &config.Config{CORSAllowedOrigins: []string{"http://localhost:5173"}}
	e := initEcho(cfg)

	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	req.Header.Set(echo.HeaderOrigin, "http://localhost:5173")
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.Contains(t, rec.Header().Get("Access-Control-Expose-Headers"), "X-Bookleaf-Maintenance")
}
