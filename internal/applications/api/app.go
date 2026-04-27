package api

import (
	"fmt"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	echoSwagger "github.com/swaggo/echo-swagger"
	"gorm.io/gorm"

	"github.com/edalferes/monetics/internal/config"
	"github.com/edalferes/monetics/internal/infra/db"
	"github.com/edalferes/monetics/internal/infra/security"
	"github.com/edalferes/monetics/internal/infra/validator"
	"github.com/edalferes/monetics/internal/modules/auth"
	"github.com/edalferes/monetics/internal/modules/budget"
	"github.com/edalferes/monetics/pkg/logger"
)

type App struct {
	echo   *echo.Echo
	db     *gorm.DB
	logger logger.Logger
	config *config.Config
}

func NewApp(cfg *config.Config) *App {
	appLogger := initLogger(cfg)
	database := initDatabase(cfg, appLogger)
	migrateDatabase(database, appLogger)
	seedDatabase(database, appLogger, cfg)
	e := initEcho()

	return &App{
		echo:   e,
		db:     database,
		logger: appLogger,
		config: cfg,
	}
}

// initLogger configures and creates the application logger
func initLogger(cfg *config.Config) logger.Logger {
	loggerConfig := logger.DefaultConfig()
	loggerConfig.Level = cfg.Logger.Level
	loggerConfig.Format = cfg.Logger.Format
	loggerConfig.Service = cfg.App.Name
	return logger.New(loggerConfig)
}

// initDatabase connects to the database
func initDatabase(cfg *config.Config, log logger.Logger) *gorm.DB {
	database, err := db.NewGormDB(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to database")
	}
	return database
}

// migrateDatabase runs migrations for all modules
func migrateDatabase(database *gorm.DB, log logger.Logger) {
	entities := []interface{}{}
	entities = append(entities, auth.Entities()...)
	entities = append(entities, budget.Entities()...)

	if err := database.AutoMigrate(entities...); err != nil {
		log.Fatal().Err(err).Msg("failed to migrate database")
	}
	log.Info().Msg("Database migration completed")
}

// seedDatabase seeds all modules
func seedDatabase(database *gorm.DB, log logger.Logger, cfg *config.Config) {
	// Seed auth module
	if err := auth.Seed(database, cfg.Admin.Username, cfg.Admin.Password); err != nil {
		log.Fatal().Err(err).Msg("failed to seed auth module")
	}

	// Seed budget module
	var adminUser struct{ ID uint }
	if err := database.Table("users").Select("id").Where("username = ?", cfg.Admin.Username).First(&adminUser).Error; err != nil {
		log.Fatal().Err(err).Msg("failed to find admin user for budget seed")
	}
	if err := budget.Seed(database, adminUser.ID); err != nil {
		log.Fatal().Err(err).Msg("failed to seed budget module")
	}
	log.Info().Msg("Database seeding completed")
}

// initEcho creates and configures Echo instance
func initEcho() *echo.Echo {
	e := echo.New()
	e.Validator = validator.NewValidator()

	// Hide Echo banner and port
	e.HideBanner = true
	e.HidePort = true

	// Security headers (RN-S1)
	e.Use(security.SecureHeaders())

	// Per-IP rate limit (RN-S2): default 50 req/s with burst of 100.
	e.Use(security.RateLimiter(security.RateLimiterConfig{
		RequestsPerSecond: 50,
		Burst:             100,
	}))

	// Configure CORS middleware
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{echo.GET, echo.POST, echo.PUT, echo.DELETE, echo.PATCH, echo.OPTIONS},
		AllowHeaders:     []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
		AllowCredentials: true,
		MaxAge:           3600,
	}))

	return e
}

// RegisterModules registers all modules
func (a *App) RegisterModules() {
	v1 := a.echo.Group("/v1")

	// Register auth module (no dependencies)
	auth.WireUp(v1, a.db, a.config, a.logger)

	// Register budget module (depends on auth)
	budget.WireUp(v1, a.db, a.config, a.logger)
}

func (a *App) RegisterGlobalRoutes() {
	// K8s probes
	a.echo.GET("/health", a.LivenessHandler)
	a.echo.GET("/ready", a.ReadinessHandler)

	// Swagger docs
	a.echo.GET("/swagger/*", echoSwagger.WrapHandler)
}

func (a *App) Run() {
	a.RegisterGlobalRoutes()
	a.RegisterModules()
	a.logger.Info().
		Int("port", a.config.App.Port).
		Str("env", a.config.App.Environment).
		Msg("Starting Server")
	a.echo.Logger.Fatal(a.echo.Start(fmt.Sprintf(":%d", a.config.App.Port)))
}
