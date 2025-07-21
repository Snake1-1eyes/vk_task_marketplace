package bootstrap

import (
	"context"
	"database/sql"
	"log"

	app_errors "github.com/Snake1-1eyes/vk_task_marketplace/internal/app_errors"
	"github.com/Snake1-1eyes/vk_task_marketplace/internal/auth"
	authRepo "github.com/Snake1-1eyes/vk_task_marketplace/internal/auth/repo/postgres"
	authUC "github.com/Snake1-1eyes/vk_task_marketplace/internal/auth/usecase"
	"github.com/Snake1-1eyes/vk_task_marketplace/internal/config"
	"github.com/Snake1-1eyes/vk_task_marketplace/internal/listing"
	listingRepo "github.com/Snake1-1eyes/vk_task_marketplace/internal/listing/repo/postgres"
	listingUC "github.com/Snake1-1eyes/vk_task_marketplace/internal/listing/usecase"
	"github.com/Snake1-1eyes/vk_task_marketplace/internal/logger"
	"github.com/Snake1-1eyes/vk_task_marketplace/internal/utils"
	postgresstorage "github.com/Snake1-1eyes/vk_task_marketplace/pkg/postgres_storage"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"go.uber.org/zap"
)

// Services содержит все сервисы приложения
type Services struct {
	AuthUseCase     auth.UseCase
	ListingsUseCase listing.UseCase
}

// Repositories содержит все репозитории приложения
type Repositories struct {
	AuthRepo     auth.Repository
	ListingsRepo listing.Repository
}

// InitializeConfig загружает конфигурацию приложения
func InitializeConfig() *config.Config {
	cfg, err := config.New()
	if err != nil {
		log.Fatalf("Ошибка загрузки конфигурации: %s", err)
	}
	return cfg
}

// InitializeLogger создает и настраивает логгер
func InitializeLogger(cfg *config.Config) *logger.Logger {
	appLogger, err := logger.New(cfg.Environment, cfg.LogLevel)
	if err != nil {
		log.Fatalf("Не удалось создать логгер: %s", err)
	}

	appLogger.Info(context.TODO(), "Логгер инициализирован",
		zap.String("environment", cfg.Environment),
		zap.String("log_level", cfg.LogLevel))

	return appLogger
}

// Bootstrap инициализирует основные компоненты приложения
func Bootstrap() (*config.Config, *logger.Logger) {
	cfg := InitializeConfig()
	appLogger := InitializeLogger(cfg)
	return cfg, appLogger
}

// InitializePostgres инициализирует соединение с PostgreSQL
func InitializePostgres(ctx context.Context, cfg *config.Config, log *logger.Logger) (*pgxpool.Pool, error) {
	pool, err := NewPool(ctx, cfg, log)
	if err != nil {
		return nil, app_errors.WrapError(err, "ошибка подключения к PostgreSQL")
	}
	return pool, nil
}

// InitializeRepositories инициализирует репозитории
func InitializeRepositories(ctx context.Context, cfg *config.Config, log *logger.Logger) (*Repositories, error) {
	pool, err := InitializePostgres(ctx, cfg, log)
	if err != nil {
		return nil, err
	}

	dbClient := postgresstorage.NewDBClient(pool, log)
	txManager := dbClient

	authRepository := authRepo.New(dbClient, txManager, log)
	listingsRepository := listingRepo.New(dbClient, txManager, log)

	return &Repositories{
		AuthRepo:     authRepository,
		ListingsRepo: listingsRepository,
	}, nil
}

// InitializeServices инициализирует сервисы (use cases)
func InitializeServices(ctx context.Context, cfg *config.Config, repos *Repositories, log *logger.Logger) *Services {
	jwtConfig := utils.JWTConfig{
		SecretKey:     cfg.JWT.SecretKey,
		TokenDuration: cfg.JWT.TokenDuration,
	}

	authService := authUC.New(repos.AuthRepo, jwtConfig, log)
	listingsService := listingUC.New(repos.ListingsRepo, log)

	return &Services{
		AuthUseCase:     authService,
		ListingsUseCase: listingsService,
	}
}

// RunMigrations запускает миграции для PostgreSQL
func RunMigrations(ctx context.Context, cfg *config.Config, log *logger.Logger) error {
	dsn := cfg.GetPostgresDSN()

	log.Info(ctx, "Запуск миграций базы данных",
		zap.String("database", cfg.Postgres.DB),
		zap.String("host", cfg.Postgres.Host))

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return app_errors.WrapError(err, "не удалось подключиться к базе данных для миграций")
	}
	defer db.Close()

	if err := db.PingContext(ctx); err != nil {
		return app_errors.WrapError(err, "не удалось проверить соединение с базой данных")
	}

	migrationsDir := cfg.Migrations.Dir

	if err := goose.Up(db, migrationsDir); err != nil {
		return app_errors.WrapError(err, "ошибка выполнения миграций")
	}

	currentVersion, err := goose.GetDBVersion(db)
	if err != nil {
		log.Warn(ctx, "Не удалось получить текущую версию миграции", zap.Error(err))
	} else {
		log.Info(ctx, "Миграции успешно выполнены", zap.Int64("version", currentVersion))
	}

	return nil
}
