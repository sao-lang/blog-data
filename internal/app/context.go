package app

import "github.com/gin-gonic/gin"

// AppContext = 你的“业务上下文”
type AppContext struct {
	*gin.Context
}

// 业务 handler 的统一签名
type Handler func(*AppContext) (any, error)
