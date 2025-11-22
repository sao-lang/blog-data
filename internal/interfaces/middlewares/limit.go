package middlewares

import (
	"time"

	"github.com/gin-gonic/gin"
)

var requestsCount = 0
var lastResetTime = time.Now()

func Limit(limit int, duration time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查是否需要重置计数器
		now := time.Now()
		if now.Sub(lastResetTime) >= duration {
			requestsCount = 0
			lastResetTime = now
		}

		// 检查请求次数是否超过限制
		if requestsCount >= limit {
			c.AbortWithStatusJSON(429, gin.H{
				"error": "Too Many Requests",
			})
			return
		}

		// 更新请求计数器
		requestsCount++

		// 执行下一个处理程序
		c.Next()
	}
}
