package main

import (
	"fmt"
	"net/http"

	"github.com/devi/bookleaf/internal/config"
	httphandler "github.com/devi/bookleaf/internal/handler"
	authmiddleware "github.com/devi/bookleaf/internal/middleware"
	"github.com/devi/bookleaf/internal/repository"
	"github.com/devi/bookleaf/internal/usecase"
	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	e := echo.New()
	e.Use(echomiddleware.Logger())
	e.Use(echomiddleware.Recover())

	cfg, err := config.Load()
	if err != nil {
		e.Logger.Fatal(err)
	}

	db, err := gorm.Open(postgres.Open(cfg.DB.URL), &gorm.Config{})
	if err != nil {
		e.Logger.Fatal(fmt.Errorf("open database connection: %w", err))
	}

	userRepository := repository.NewUserRepository(db)
	userUsecase := usecase.NewUserUsecase(userRepository)
	meHandler := httphandler.NewMeHandler(userUsecase)

	authMiddleware, err := authmiddleware.NewAuthMiddleware(cfg.Kinde.IssuerURL, cfg.Kinde.Audience, userUsecase)
	if err != nil {
		e.Logger.Fatal(fmt.Errorf("initialise auth middleware: %w", err))
	}

	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	protected := e.Group("")
	protected.Use(authMiddleware)
	protected.GET("/me", meHandler.GetMe)

	e.Logger.Fatal(e.Start(":" + cfg.Port))
}
