package app

import (
	"blog/internal/config"
	"blog/internal/infra/minio"
	"blog/internal/infra/pgsql"
	"blog/internal/infra/redis"
	"blog/internal/interfaces/middlewares"
	router "blog/internal/router"
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

func Setup() (*gin.Engine, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, err
	}
	pgsql, err := pgsql.NewPGSQL(loadPgsqlConfig(cfg))
	if err != nil {
		return nil, err
	}
	engin := gin.Default()
	registerMiddlewares(engin)
	router.SetupRouter(engin, pgsql.DB)
	app := App{
		Pgsql:  pgsql,
		Router: engin,
	}
	return app.Router, nil
}

func registerMiddlewares(engin *gin.Engine) {
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "dev"
	}
	engin.Use(middlewares.Response())
	engin.Use(middlewares.Logger(env))
	engin.Use(middlewares.CORS())
	engin.Use(middlewares.Recovery())
}
