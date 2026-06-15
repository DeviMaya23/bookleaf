package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	httphandler "github.com/devi/bookleaf/internal/handler"
	authmiddleware "github.com/devi/bookleaf/internal/handler/middleware"
	"github.com/devi/bookleaf/internal/kinde"
	"github.com/devi/bookleaf/internal/platform/config"
	"github.com/devi/bookleaf/internal/platform/observability"
	"github.com/devi/bookleaf/internal/repository"
	"github.com/devi/bookleaf/internal/storage"
	"github.com/devi/bookleaf/internal/usecase"
	"github.com/devi/bookleaf/internal/vision"
	"github.com/devi/bookleaf/internal/worker"
	pgx "github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	otelgorm "gorm.io/plugin/opentelemetry/tracing"
)

// riverEnqueuer adapts *river.Client to usecase.JobEnqueuer.
type riverEnqueuer struct {
	client *river.Client[pgx.Tx]
}

func (e *riverEnqueuer) Insert(ctx context.Context, args usecase.JobArgs) error {
	riverArgs, ok := args.(river.JobArgs)
	if !ok {
		return fmt.Errorf("unsupported job args type %T", args)
	}
	_, err := e.client.Insert(ctx, riverArgs, &river.InsertOpts{MaxAttempts: args.MaxAttempts()})
	return err
}

type server struct {
	echo        *echo.Echo
	riverClient *river.Client[pgx.Tx]
	riverPool   *pgxpool.Pool
	logger      *zap.Logger
	shutdownTel func(context.Context)
}

func newServer(ctx context.Context, cfg *config.Config, logger *zap.Logger) *server {
	e := initEcho(cfg)
	tel, shutdownTel := initTelemetry(ctx, cfg, e, logger)
	db := initDB(cfg, logger)
	riverClient, riverPool := initApp(ctx, cfg, db, tel, e, logger)
	return &server{
		echo:        e,
		riverClient: riverClient,
		riverPool:   riverPool,
		logger:      logger,
		shutdownTel: shutdownTel,
	}
}

func (s *server) start(port string) error {
	return s.echo.Start(":" + port)
}

func (s *server) shutdown(ctx context.Context) {
	if err := s.riverClient.Stop(ctx); err != nil {
		s.logger.Error("river client stop", zap.Error(err))
	}
	s.riverPool.Close()
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
			"X-Bookleaf-Bypass",
		},
		ExposeHeaders: []string{
			"X-Bookleaf-Maintenance",
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

func initApp(ctx context.Context, cfg *config.Config, db *gorm.DB, tel *observability.Telemetry, e *echo.Echo, logger *zap.Logger) (*river.Client[pgx.Tx], *pgxpool.Pool) {
	userRepository := repository.NewUserRepository(db)
	userUsecase := usecase.NewUserUsecase(userRepository, tel)
	folderRepository := repository.NewFolderRepository(db)
	storageService := storage.NewR2Storage(cfg.R2, tel)
	imageRepository := repository.NewImageRepository(db)
	pendingUploadRepository := repository.NewPendingUploadRepository(db)
	tagRepository := repository.NewTagRepository(db)
	folderUsecase := usecase.NewFolderUsecase(folderRepository, imageRepository, storageService, tel)
	tagUsecase := usecase.NewTagUsecase(tagRepository, tel)

	var visionService usecase.VisionService
	if cfg.Vision.APIKey != "" {
		visionService = vision.NewVisionClient(cfg.Vision.APIKey)
	}

	// Dedicated River pool (max 3 connections, separate from GORM's pool).
	poolCfg, err := pgxpool.ParseConfig(cfg.DB.URL)
	if err != nil {
		logger.Fatal("parse river pool config", zap.Error(err))
	}
	poolCfg.MaxConns = 3
	riverPool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		logger.Fatal("open river database pool", zap.Error(err))
	}

	// Deferred enqueuer: client field is set after river.NewClient to break the
	// uploadUsecase ↔ riverClient init cycle.
	enqueuer := &riverEnqueuer{}

	imageUsecase := usecase.NewImageUsecase(imageRepository, tagRepository, folderRepository, storageService, tel)
	trashUsecase := usecase.NewTrashUsecase(imageRepository, storageService, enqueuer, tel)
	folderShareRepository := repository.NewFolderShareRepository(db)
	shareUsecase := usecase.NewShareUsecase(folderShareRepository, folderRepository, imageRepository, storageService, tel)
	uploadUsecase := usecase.NewImageUploadUsecase(imageRepository, pendingUploadRepository, folderRepository, userRepository, storageService, visionService, enqueuer, tel)

	accountRepository := repository.NewAccountRepository(db)
	kindeClient := kinde.NewClient(cfg.Kinde)
	accountUsecase := usecase.NewAccountUsecase(accountRepository, userRepository, kindeClient, enqueuer, tel)

	workers := river.NewWorkers()
	river.AddWorker(workers, worker.NewVisionWorker(uploadUsecase))
	river.AddWorker(workers, worker.NewCleanupStaleUploadsWorker(uploadUsecase))
	river.AddWorker(workers, worker.NewTrashPurgeWorker(trashUsecase))
	river.AddWorker(workers, worker.NewR2DeleteWorker(trashUsecase))
	river.AddWorker(workers, worker.NewAccountKindeDeletionWorker(accountUsecase))
	river.AddWorker(workers, worker.NewAccountKindeDeletionReconcileWorker(accountUsecase))

	riverClient, err := river.NewClient(riverpgxv5.New(riverPool), &river.Config{
		Queues: map[string]river.QueueConfig{
			river.QueueDefault: {MaxWorkers: 10},
		},
		Workers: workers,
		PeriodicJobs: []*river.PeriodicJob{
			river.NewPeriodicJob(
				river.PeriodicInterval(10*time.Minute),
				func() (river.JobArgs, *river.InsertOpts) { return worker.CleanupStaleUploadsArgs{}, nil },
				nil,
			),
			river.NewPeriodicJob(
				river.PeriodicInterval(24*time.Hour),
				func() (river.JobArgs, *river.InsertOpts) { return worker.TrashPurgeArgs{}, nil },
				nil,
			),
			river.NewPeriodicJob(
				river.PeriodicInterval(24*time.Hour),
				func() (river.JobArgs, *river.InsertOpts) { return worker.AccountKindeDeletionReconcileArgs{}, nil },
				nil,
			),
		},
	})
	if err != nil {
		logger.Fatal("create river client", zap.Error(err))
	}

	enqueuer.client = riverClient

	if err := riverClient.Start(ctx); err != nil {
		logger.Fatal("start river client", zap.Error(err))
	}

	authMiddleware, err := authmiddleware.NewAuthMiddleware(cfg.Kinde.IssuerURL, cfg.Kinde.Audience, userUsecase, logger)
	if err != nil {
		logger.Fatal("initialise auth middleware", zap.Error(err))
	}

	meHandler := httphandler.NewMeHandler(userUsecase, accountUsecase, tel)
	folderHandler := httphandler.NewFolderHandler(folderUsecase, tel)
	tagHandler := httphandler.NewTagHandler(tagUsecase, tel)
	imageHandler := httphandler.NewImageHandler(imageUsecase, tel)
	trashHandler := httphandler.NewTrashHandler(trashUsecase, tel)
	shareHandler := httphandler.NewShareHandler(shareUsecase, tel)
	uploadHandler := httphandler.NewUploadHandler(uploadUsecase, tel)
	healthHandler := httphandler.NewHealthHandler(db, storageService)

	e.GET("/health", healthHandler.GetHealth)
	e.GET("/share/:token", shareHandler.GetSharedFolder)

	protected := e.Group("")
	protected.Use(authmiddleware.NewMaintenanceMiddleware(cfg.Maintenance))
	protected.Use(authMiddleware)
	protected.Use(observability.LoggingMiddleware(tel, authmiddleware.AuthenticatedUserIDFromContext))
	protected.GET("/me", meHandler.GetMe)
	protected.PATCH("/me", meHandler.UpdateMe)
	protected.DELETE("/me", meHandler.DeleteMe)
	protected.POST("/folders", folderHandler.CreateFolder)
	protected.GET("/folders", folderHandler.ListFolders)
	protected.GET("/folders/:id", folderHandler.GetFolder)
	protected.GET("/folders/:id/export", folderHandler.ExportFolder)
	protected.PATCH("/folders/:id", folderHandler.UpdateFolder)
	protected.DELETE("/folders/:id", folderHandler.DeleteFolder)
	protected.POST("/folders/:id/share", shareHandler.CreateShare)
	protected.GET("/folders/:id/share", shareHandler.GetShare)
	protected.DELETE("/folders/:id/share", shareHandler.DeleteShare)
	protected.POST("/tags", tagHandler.CreateTag)
	protected.GET("/tags", tagHandler.ListTags)
	protected.PUT("/tags/:id", tagHandler.UpdateTag)
	protected.DELETE("/tags/:id", tagHandler.DeleteTag)
	protected.POST("/images", uploadHandler.InitiateUpload)
	protected.POST("/images/:id/complete", uploadHandler.CompleteUpload)
	protected.POST("/images/:id/accept-suggestion", uploadHandler.AcceptSuggestion)
	protected.GET("/images/trash", trashHandler.ListTrashed)
	protected.DELETE("/images/trash", trashHandler.EmptyTrash)
	protected.DELETE("/images/trash/:id", trashHandler.DeleteFromTrash)
	protected.GET("/images", imageHandler.ListImages)
	protected.GET("/images/in-folder/:id", imageHandler.ListFolderImages)
	protected.GET("/images/:id", imageHandler.GetImage)
	protected.GET("/images/:id/download", imageHandler.DownloadImage)
	protected.POST("/images/:id/move-folder", imageHandler.MoveImageFolder)
	protected.PATCH("/images/:id/position", imageHandler.UpdateImagePosition)
	protected.PATCH("/images/:id", imageHandler.UpdateImage)
	protected.DELETE("/images/:id", trashHandler.SoftDelete)
	protected.POST("/images/:id/restore", trashHandler.Restore)

	return riverClient, riverPool
}
