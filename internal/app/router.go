package app

import (
	"blog/internal/domain/user"
	"blog/internal/interfaces/handlers"
	"blog/internal/interfaces/middlewares"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func setupRouter(router *gin.Engine, db *gorm.DB) {
	registerMiddlewares(router)
	setupUserRouter(router, db)
}

func setupUserRouter(router *gin.Engine, db *gorm.DB) {
	userRepository := user.NewUserRepository(db)
	userService := user.NewUserService(userRepository)
	userHandler := handlers.NewUserHandler(userService)

	auth := router.Group("/auth")
	{
		auth.POST("/register", middlewares.Validate(&user.CreateUserDTO{}), userHandler.Register)
		auth.POST("/login", middlewares.Auth(), middlewares.Validate(&user.CreateUserDTO{}), userHandler.Login)
	}
}

func registerMiddlewares(router *gin.Engine) {
	router.Use(middlewares.Response())
	router.Use(middlewares.Logger())
	router.Use(middlewares.CORS())
}
