package router

import (
	"blog/internal/domain/user"
	"blog/internal/infra/gnest"
	"blog/internal/interfaces/handlers"
	"blog/internal/interfaces/middlewares"
)

func setupAuthRouter(app *gnest.GnestApp) {
	app.Provide(
		&user.UserRepository{},
		&user.UserService{},
	)

	// 3. 注册控制器
	userCtrl := &handlers.UserController{}
	app.Provide(userCtrl)

	// 4. 声明路由 (替代原来的 router 文件夹功能)
	auth := app.Group("/auth")
	{
		// 注意：这里不需要再传 middlewares.Validate，gnest 内部已包含自动校验
		auth.POST("/register", userCtrl.Register)

		// 鉴权中间件可以继续用
		auth.POST("/login", userCtrl.Login, middlewares.Auth())

		auth.POST("/refresh-token", userCtrl.RefreshToken)
	}
}
