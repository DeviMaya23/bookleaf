package handler

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/devi/bookleaf/internal/storage"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

const healthProbeTimeout = 3 * time.Second

type HealthHandler struct {
	db    *gorm.DB
	store storage.StorageService

	dbProbe func(ctx context.Context) error
	r2Probe func(ctx context.Context) error
}

type healthResponse struct {
	Status string `json:"status"`
	DB     string `json:"db"`
	R2     string `json:"r2"`
}

func NewHealthHandler(db *gorm.DB, store storage.StorageService) *HealthHandler {
	h := &HealthHandler{
		db:    db,
		store: store,
	}

	h.dbProbe = func(ctx context.Context) error {
		if h.db == nil {
			return errors.New("db is not configured")
		}
		return h.db.WithContext(ctx).Exec("SELECT 1").Error
	}

	h.r2Probe = func(ctx context.Context) error {
		if h.store == nil {
			return errors.New("r2 storage is not configured")
		}
		return h.store.Ping(ctx)
	}

	return h
}

func (h *HealthHandler) GetHealth(c echo.Context) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), healthProbeTimeout)
	defer cancel()

	res := healthResponse{
		Status: "ok",
		DB:     "ok",
		R2:     "ok",
	}

	if err := h.dbProbe(ctx); err != nil {
		res.Status = "degraded"
		res.DB = err.Error()
	}

	if err := h.r2Probe(ctx); err != nil {
		res.Status = "degraded"
		res.R2 = err.Error()
	}

	return c.JSON(http.StatusOK, res)
}
