package app

import (
	"blog/internal/config"
	"blog/internal/infra/minio"
	"blog/internal/infra/pgsql"
	"blog/internal/infra/redis"

	"github.com/gin-gonic/gin"
)

type App[T any] struct {
	Pgsql  *pgsql.PGSQL[T]
	Minio  *minio.Client
	Redis  *redis.Client
	Router *gin.Engine
}

func Setup[T any]() (*gin.Engine, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, err
	}
	pgsql, err := pgsql.NewPGSQL[T](pgsql.Config{
		Host:     cfg.PgSQL.Host,
		Port:     cfg.PgSQL.Port,
		User:     cfg.PgSQL.User,
		Password: cfg.PgSQL.Password,
		DBName:   cfg.PgSQL.DBName,
		SSLMode:  cfg.PgSQL.SSLMode,
		MaxIdle:  cfg.PgSQL.MaxIdle,
		MaxOpen:  cfg.PgSQL.MaxOpen,
		LogLevel: cfg.PgSQL.LogLevel,
	})
	if err != nil {
		return nil, err
	}
	router := gin.Default()
	setupRouter(router, pgsql.DB)
	app := App[T]{
		Pgsql:  pgsql,
		Router: router,
	}
	return app.Router, nil
}
