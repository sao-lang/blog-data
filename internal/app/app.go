package app

import (
	"blog/internal/config"
	"blog/internal/infra/gnest"
	"blog/internal/infra/logger"
	"blog/internal/infra/minio"
	"blog/internal/infra/pgsql"
	"blog/internal/infra/redis"
	"blog/internal/interfaces/interceptors"
	"blog/internal/interfaces/middlewares"
	"blog/internal/router"
	"os"

	"github.com/gin-gonic/gin"
)

type App struct {
	Pgsql  *pgsql.PGSQL
	Minio  *minio.Client
	Redis  *redis.Client
	Router *gin.Engine
}

func loadPgsqlConfig(cfg *config.Config) pgsql.Config {
	return pgsql.Config{
		Host:     cfg.PgSQL.Host,
		Port:     cfg.PgSQL.Port,
		User:     cfg.PgSQL.User,
		Password: cfg.PgSQL.Password,
		DBName:   cfg.PgSQL.DBName,
		SSLMode:  cfg.PgSQL.SSLMode,
		MaxIdle:  cfg.PgSQL.MaxIdle,
		MaxOpen:  cfg.PgSQL.MaxOpen,
		LogLevel: cfg.PgSQL.LogLevel,
	}
}

func Setup() (*gnest.GnestApp, error) {
	app := gnest.New()
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, err
	}
	// 1. 提供底层依赖 (注入到容器)
	pg, err := pgsql.NewPGSQL(loadPgsqlConfig(cfg))
	if err != nil {
		return nil, err
	}
	app.Provide(pg.DB) // 提供 *gorm.DB
	router.Setup(app)
	registerMiddlewares(app)
	app.GET("/string", func(ctx *gin.Context) string {
		return "This is a direct string response from Gnest!"
	})

	return app, nil
}

func registerMiddlewares(app *gnest.GnestApp) {
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "dev"
	}

	app.Provide(logger.NewLoggerService(env))
	logInterceptor := &interceptors.LoggingInterceptor{}
	app.Provide(logInterceptor)
	app.UseGlobalInterceptors(logInterceptor)
	// app.Use(middlewares.Response())
	// app.Use(middlewares.Logger(env))
	app.Use(middlewares.CORS())
	app.Use(middlewares.Recovery())
}
