package main

import (
	"fmt"
	"net/http"
	"os"

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

	kindeIssuerURL, err := requireEnv("KINDE_ISSUER_URL")
	if err != nil {
		e.Logger.Fatal(err)
	}

	kindeAudience, err := requireEnv("KINDE_AUDIENCE")
	if err != nil {
		e.Logger.Fatal(err)
	}

	databaseURL, err := requireEnv("DATABASE_URL")
	if err != nil {
		e.Logger.Fatal(err)
	}

	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{})
	if err != nil {
		e.Logger.Fatal(fmt.Errorf("open database connection: %w", err))
	}

	userRepository := repository.NewUserRepository(db)
	userUsecase := usecase.NewUserUsecase(userRepository)
	meHandler := httphandler.NewMeHandler(userUsecase)

	authMiddleware, err := authmiddleware.NewAuthMiddleware(kindeIssuerURL, kindeAudience, userUsecase)
	if err != nil {
		e.Logger.Fatal(fmt.Errorf("initialise auth middleware: %w", err))
	}

	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	protected := e.Group("")
	protected.Use(authMiddleware)
	protected.GET("/me", meHandler.GetMe)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	e.Logger.Fatal(e.Start(":" + port))
}

func requireEnv(name string) (string, error) {
	value := os.Getenv(name)
	if value == "" {
		return "", fmt.Errorf("%s is required", name)
	}
	return value, nil
}
