package router

import (
	"blog/internal/infra/gnest"
)

func Setup(app *gnest.GnestApp) {

	setupAuthRouter(app)
}
