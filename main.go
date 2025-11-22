package main

import (
	"blog/internal/config"
	"blog/internal/dto"
	"blog/internal/handlers"
	"blog/internal/middlewares"
	"blog/internal/repositories"
	"blog/internal/services"
	"blog/internal/utils"
	"fmt"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	// "context"
	// "net/http"
	// "os"
)

func setupConfig() (*config.Config, error) {
	conf, err := config.LoadConfig()
	if err != nil {
		return nil, err
	}
	return conf, nil
}

func registerMiddlewares(router *gin.Engine) {
	router.Use(middlewares.Response())
	router.Use(middlewares.Logger())
	router.Use(middlewares.CORS())
}

func setupUserRouter(router *gin.Engine, db *gorm.DB) {
	userRepository := repositories.NewUserRepository(db)
	userService := services.NewUserService(userRepository)
	userHandler := handlers.NewAuthHandler(userService)

	auth := router.Group("/auth")
	{
		auth.POST("/register", middlewares.Validate(&dto.CreateUserDTO{}), userHandler.Register)
		auth.POST("/login", middlewares.Auth(), middlewares.Validate(&dto.CreateUserDTO{}), userHandler.Login)
	}
}

func setupRouter(db *gorm.DB) *gin.Engine {
	router := gin.Default()
	registerMiddlewares(router)
	router.GET("/songs", handlers.GetSongs)

	setupUserRouter(router, db)
	return router
}

func main() {
	globalConfig, err := setupConfig()
	if err != nil {
		panic(fmt.Errorf("config load failed: %w", err))
	}
	db, err := setupPgSql(globalConfig)
	if err != nil {
		panic(fmt.Errorf("pgsql connect failed: %w", err))
	}
	// setupMinIO()
	port := utils.FindAvailablePort(8089)
	router := setupRouter(db)
	router.Run(fmt.Sprintf("0.0.0.0:%d", port))
}
