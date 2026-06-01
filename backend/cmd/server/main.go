package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/devi/bookleaf/internal/config"
	httphandler "github.com/devi/bookleaf/internal/handler"
	authmiddleware "github.com/devi/bookleaf/internal/middleware"
	"github.com/devi/bookleaf/internal/observability"
	"github.com/devi/bookleaf/internal/repository"
	"github.com/devi/bookleaf/internal/storage"
	"github.com/devi/bookleaf/internal/thumbnail"
	"github.com/devi/bookleaf/internal/usecase"
	"github.com/devi/bookleaf/internal/vision"
	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	otelgorm "gorm.io/plugin/opentelemetry/tracing"
)

type server struct {
	echo         *echo.Echo
	imageUsecase usecase.ImageUsecase
	logger       *zap.Logger
	shutdownTel  func(context.Context)
}

func newServer(ctx context.Context, cfg *config.Config, logger *zap.Logger) *server {
	e := initEcho(cfg)
	tel, shutdownTel := initTelemetry(ctx, cfg, e, logger)
	db := initDB(cfg, logger)
	imageUsecase := initApp(cfg, db, tel, e, logger)
	return &server{
		echo:        e,
		imageUsecase: imageUsecase,
		logger:      logger,
		shutdownTel: shutdownTel,
	}
}

func (s *server) startWorkers(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(10 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			if err := s.imageUsecase.CleanupStaleUploads(ctx, 30*time.Minute); err != nil {
				s.logger.Warn("stale upload cleanup failed", zap.Error(err))
			}
		}
	}()

	go func() {
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			if err := s.imageUsecase.PurgeExpiredTrash(ctx, 30*24*time.Hour); err != nil {
				s.logger.Warn("trash purge failed", zap.Error(err))
			}
		}
	}()
}

func (s *server) start(port string) error {
	return s.echo.Start(":" + port)
}

func (s *server) shutdown(ctx context.Context) {
	s.shutdownTel(ctx)
	if err := s.echo.Shutdown(ctx); err != nil {
		s.logger.Error("echo shutdown", zap.Error(err))
	}
	_ = s.logger.Sync()
}

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

	srv := newServer(ctx, cfg, logger)
	srv.startWorkers(ctx)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		<-quit
		srv.logger.Info("shutting down")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		srv.shutdown(shutdownCtx)
	}()

	if err := srv.start(cfg.Port); err != nil && err != http.ErrServerClosed {
		logger.Fatal("server stopped", zap.Error(err))
	}
}

func initEcho(cfg *config.Config) *echo.Echo {
	e := echo.New()
	e.Use(echomiddleware.Recover())
	e.Use(echomiddleware.CORSWithConfig(echomiddleware.CORSConfig{
		AllowOrigins: cfg.CORSAllowedOrigins,
		AllowHeaders: []string{
			echo.HeaderAuthorization,
			echo.HeaderContentType,
		},
	}))
	return e
}

func initTelemetry(ctx context.Context, cfg *config.Config, e *echo.Echo, logger *zap.Logger) (*observability.Telemetry, func(context.Context)) {
	otel.SetErrorHandler(otel.ErrorHandlerFunc(func(err error) {
		logger.Error("otel error", zap.Error(err))
	}))

	if !cfg.Obs.OTELEnabled {
		return observability.NewTelemetry(logger, nil, nil), func(context.Context) {}
	}

	tp, err := observability.NewTracerProvider(ctx, cfg.Obs.OTELExporter, cfg.Obs.SampleRatio)
	if err != nil {
		logger.Fatal("init tracer provider", zap.Error(err))
	}

	mp, metricsHandler, err := observability.NewMeterProvider(cfg.Obs.OTELMetricsExporter)
	if err != nil {
		logger.Fatal("init meter provider", zap.Error(err))
	}

	tel := observability.NewTelemetry(logger, otel.Tracer("bookleaf"), otel.Meter("bookleaf"))
	e.Use(observability.TraceMiddleware(otel.Tracer("bookleaf")))
	e.Use(observability.MetricsMiddleware(otel.Meter("bookleaf")))
	if metricsHandler != nil {
		e.GET("/metrics", echo.WrapHandler(metricsHandler))
	}

	shutdown := func(ctx context.Context) {
		if err := tp.Shutdown(ctx); err != nil {
			logger.Error("tracer shutdown", zap.Error(err))
		}
		if err := mp.Shutdown(ctx); err != nil {
			logger.Error("meter shutdown", zap.Error(err))
		}
	}

	return tel, shutdown
}

func initDB(cfg *config.Config, logger *zap.Logger) *gorm.DB {
	db, err := gorm.Open(postgres.Open(cfg.DB.URL), &gorm.Config{
		Logger: repository.NewZapGORMLogger(logger),
	})
	if err != nil {
		logger.Fatal("open database connection", zap.Error(err))
	}

	sqlDB, err := db.DB()
	if err != nil {
		logger.Fatal("get underlying sql.DB", zap.Error(err))
	}
	sqlDB.SetMaxOpenConns(5)
	sqlDB.SetMaxIdleConns(3)
	sqlDB.SetConnMaxLifetime(15 * time.Minute)
	sqlDB.SetConnMaxIdleTime(5 * time.Minute)

	if cfg.Obs.OTELEnabled {
		if err := db.Use(otelgorm.NewPlugin()); err != nil {
			logger.Fatal("register otelgorm plugin", zap.Error(err))
		}
	}

	return db
}

func initApp(cfg *config.Config, db *gorm.DB, tel *observability.Telemetry, e *echo.Echo, logger *zap.Logger) usecase.ImageUsecase {
	userRepository := repository.NewUserRepository(db)
	userUsecase := usecase.NewUserUsecase(userRepository, tel)
	folderRepository := repository.NewFolderRepository(db)
	storageService := storage.NewR2Storage(cfg.R2, tel)
	thumbnailService := thumbnail.NewThumbnailService()
	imageRepository := repository.NewImageRepository(db)
	pendingUploadRepository := repository.NewPendingUploadRepository(db)
	tagRepository := repository.NewTagRepository(db)
	folderUsecase := usecase.NewFolderUsecase(folderRepository, imageRepository, tel)
	tagUsecase := usecase.NewTagUsecase(tagRepository, tel)

	var visionService vision.VisionService
	if cfg.Vision.APIKey != "" {
		visionService = vision.NewVisionClient(cfg.Vision.APIKey)
	}

	imageUsecase := usecase.NewImageUsecase(imageRepository, pendingUploadRepository, tagRepository, storageService, thumbnailService, visionService, folderRepository, userRepository, tel)

	authMiddleware, err := authmiddleware.NewAuthMiddleware(cfg.Kinde.IssuerURL, cfg.Kinde.Audience, userUsecase, logger)
	if err != nil {
		logger.Fatal("initialise auth middleware", zap.Error(err))
	}

	meHandler := httphandler.NewMeHandler(userUsecase, tel)
	folderHandler := httphandler.NewFolderHandler(folderUsecase, tel)
	tagHandler := httphandler.NewTagHandler(tagUsecase, tel)
	imageHandler := httphandler.NewImageHandler(imageUsecase, tel)
	healthHandler := httphandler.NewHealthHandler(db, storageService)

	e.GET("/health", healthHandler.GetHealth)

	protected := e.Group("")
	protected.Use(authMiddleware)
	protected.Use(observability.LoggingMiddleware(tel, authmiddleware.AuthenticatedUserIDFromContext))
	protected.GET("/me", meHandler.GetMe)
	protected.POST("/folders", folderHandler.CreateFolder)
	protected.GET("/folders", folderHandler.ListFolders)
	protected.GET("/folders/:id", folderHandler.GetFolder)
	protected.PUT("/folders/:id", folderHandler.UpdateFolder)
	protected.DELETE("/folders/:id", folderHandler.DeleteFolder)
	protected.POST("/tags", tagHandler.CreateTag)
	protected.GET("/tags", tagHandler.ListTags)
	protected.PUT("/tags/:id", tagHandler.UpdateTag)
	protected.DELETE("/tags/:id", tagHandler.DeleteTag)
	protected.POST("/images", imageHandler.InitiateUpload)
	protected.POST("/images/:id/complete", imageHandler.CompleteUpload)
	protected.POST("/images/:id/accept-suggestion", imageHandler.AcceptSuggestion)
	protected.GET("/images/trash", imageHandler.ListTrashed)
	protected.GET("/images", imageHandler.ListImages)
	protected.GET("/images/:id", imageHandler.GetImage)
	protected.GET("/images/:id/download", imageHandler.DownloadImage)
	protected.POST("/images/:id/move-folder", imageHandler.MoveImageFolder)
	protected.PATCH("/images/:id/position", imageHandler.UpdateImagePosition)
	protected.PATCH("/images/:id", imageHandler.UpdateImage)
	protected.DELETE("/images/:id", imageHandler.SoftDelete)
	protected.POST("/images/:id/restore", imageHandler.Restore)

	return imageUsecase
}
