package main

import (
	"fmt"
	"net/http"

	"github.com/devi/bookleaf/internal/config"
	httphandler "github.com/devi/bookleaf/internal/handler"
	authmiddleware "github.com/devi/bookleaf/internal/middleware"
	"github.com/devi/bookleaf/internal/repository"
	"github.com/devi/bookleaf/internal/storage"
	"github.com/devi/bookleaf/internal/thumbnail"
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
	folderRepository := repository.NewFolderRepository(db)
	folderUsecase := usecase.NewFolderUsecase(folderRepository)
	folderHandler := httphandler.NewFolderHandler(folderUsecase)
	storageService := storage.NewR2Storage(cfg.R2)
	thumbnailService := thumbnail.NewThumbnailService()
	imageRepository := repository.NewImageRepository(db)
	imageUsecase := usecase.NewImageUsecase(imageRepository, storageService, thumbnailService)
	imageHandler := httphandler.NewImageHandler(imageUsecase, storageService)

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
	protected.POST("/folders", folderHandler.CreateFolder)
	protected.GET("/folders", folderHandler.ListFolders)
	protected.GET("/folders/:id", folderHandler.GetFolder)
	protected.PUT("/folders/:id", folderHandler.UpdateFolder)
	protected.DELETE("/folders/:id", folderHandler.DeleteFolder)
	protected.POST("/images", imageHandler.InitiateUpload)
	protected.POST("/images/:id/complete", imageHandler.CompleteUpload)
	protected.GET("/images/trash", imageHandler.ListTrashed)
	protected.GET("/images", imageHandler.ListImages)
	protected.GET("/images/:id", imageHandler.GetImage)
	protected.DELETE("/images/:id", imageHandler.SoftDelete)
	protected.POST("/images/:id/restore", imageHandler.Restore)

	e.Logger.Fatal(e.Start(":" + cfg.Port))
}
