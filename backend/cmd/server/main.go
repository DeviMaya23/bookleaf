package main

import (
	"context"
	"fmt"

	"github.com/devi/bookleaf/internal/config"
	httphandler "github.com/devi/bookleaf/internal/handler"
	authmiddleware "github.com/devi/bookleaf/internal/middleware"
	"github.com/devi/bookleaf/internal/observability"
	"github.com/devi/bookleaf/internal/repository"
	"github.com/devi/bookleaf/internal/storage"
	"github.com/devi/bookleaf/internal/thumbnail"
	"github.com/devi/bookleaf/internal/usecase"
	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	otelgorm "gorm.io/plugin/opentelemetry/tracing"
)

func main() {
	ctx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		panic(fmt.Errorf("load config: %w", err))
	}

	logger, err := observability.NewLogger(cfg.Obs.LogFormat)
	if err != nil {
		panic(fmt.Errorf("init logger: %w", err))
	}
	defer logger.Sync()

	tp, err := observability.NewTracerProvider(ctx, cfg.Obs.OTELExporter)
	if err != nil {
		logger.Fatal("init tracer provider", zap.Error(err))
	}
	defer tp.Shutdown(ctx)

	mp, metricsHandler, err := observability.NewMeterProvider(cfg.Obs.OTELMetricsExporter)
	if err != nil {
		logger.Fatal("init meter provider", zap.Error(err))
	}
	defer mp.Shutdown(ctx)

	tel := observability.NewTelemetry(logger, otel.Tracer("bookleaf"), otel.Meter("bookleaf"))

	e := echo.New()
	e.Use(echomiddleware.Recover())
	e.Use(observability.TraceMiddleware(otel.Tracer("bookleaf")))
	e.Use(observability.MetricsMiddleware(otel.Meter("bookleaf")))

	db, err := gorm.Open(postgres.Open(cfg.DB.URL), &gorm.Config{
		Logger: repository.NewZapGORMLogger(logger),
	})
	if err != nil {
		logger.Fatal("open database connection", zap.Error(err))
	}
	if err := db.Use(otelgorm.NewPlugin()); err != nil {
		logger.Fatal("register otelgorm plugin", zap.Error(err))
	}

	userRepository := repository.NewUserRepository(db)
	userUsecase := usecase.NewUserUsecase(userRepository, tel)
	meHandler := httphandler.NewMeHandler(userUsecase, tel)
	folderRepository := repository.NewFolderRepository(db)
	folderUsecase := usecase.NewFolderUsecase(folderRepository, tel)
	folderHandler := httphandler.NewFolderHandler(folderUsecase, tel)
	storageService := storage.NewR2Storage(cfg.R2, tel)
	thumbnailService := thumbnail.NewThumbnailService()
	imageRepository := repository.NewImageRepository(db)
	imageUsecase := usecase.NewImageUsecase(imageRepository, storageService, thumbnailService, tel)
	imageHandler := httphandler.NewImageHandler(imageUsecase, storageService, tel)

	authMiddleware, err := authmiddleware.NewAuthMiddleware(cfg.Kinde.IssuerURL, cfg.Kinde.Audience, userUsecase, logger)
	if err != nil {
		logger.Fatal("initialise auth middleware", zap.Error(err))
	}

	healthHandler := httphandler.NewHealthHandler(db, storageService)
	e.GET("/health", healthHandler.GetHealth)

	if metricsHandler != nil {
		e.GET("/metrics", echo.WrapHandler(metricsHandler))
	}

	protected := e.Group("")
	protected.Use(authMiddleware)
	protected.Use(observability.LoggingMiddleware(tel, authmiddleware.AuthenticatedUserIDFromContext))
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
	protected.PATCH("/images/:id", imageHandler.UpdateImage)
	protected.DELETE("/images/:id", imageHandler.SoftDelete)
	protected.POST("/images/:id/restore", imageHandler.Restore)

	if err := e.Start(":" + cfg.Port); err != nil {
		logger.Fatal("server stopped", zap.Error(err))
	}
}
