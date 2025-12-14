package interceptors

import (
	"blog/internal/infra/logger" // ！！！请替换为你的模块名
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// bodyLogWriter 用于捕获由 Gin 直接输出的响应体（c.JSON等）
type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

type LoggingInterceptor struct {
	// Gnest 依赖注入
	LoggerService *logger.LoggerService
}

func (l *LoggingInterceptor) Intercept(c *gin.Context, next func() interface{}) interface{} {
	start := time.Now()

	// 1. 获取请求 Headers
	reqHeaders := make(map[string]interface{})
	for k, v := range c.Request.Header {
		reqHeaders[k] = v
	}

	// 2. 捕获请求 Body
	var reqBody []byte
	if c.Request.Body != nil {
		reqBody, _ = io.ReadAll(c.Request.Body)
		// 必须重新填充 Body，否则下游无法再次读取
		c.Request.Body = io.NopCloser(bytes.NewBuffer(reqBody))
	}

	// 3. 装饰 ResponseWriter 以捕获响应
	blw := &bodyLogWriter{
		ResponseWriter: c.Writer,
		body:           &bytes.Buffer{},
	}
	c.Writer = blw

	// 执行业务逻辑（进入 Controller）
	result := next()

	// 4. 计算耗时与环境判断
	duration := time.Since(start)
	env := os.Getenv("APP_ENV")

	// 5. 捕获响应 Headers
	respHeaders := make(map[string]interface{})
	for k, v := range c.Writer.Header() {
		respHeaders[k] = v
	}

	// 6. 核心：处理响应数据 (解决 /string 为空的问题)
	var finalResp string
	if result != nil {
		// 优先从 Gnest 返回值中获取
		if s, ok := result.(string); ok {
			finalResp = s
		} else {
			// 如果是结构体或 Map，转为 JSON
			if b, err := json.Marshal(result); err == nil {
				finalResp = string(b)
			} else {
				finalResp = fmt.Sprintf("%v", result)
			}
		}
	} else {
		// 如果 Controller 返回 nil，则尝试从 Buffer 中读取直接输出的内容
		finalResp = blw.body.String()
	}

	// 7. 【控制台美化打印】
	if env != "prod" {
		reqH, _ := json.Marshal(reqHeaders)
		respH, _ := json.Marshal(respHeaders)

		fmt.Printf("\n\033[32m============ %s ============\033[0m\n", time.Now().Format("2006-01-02 15:04:05"))
		fmt.Printf("\"url\": \"%s\"\n", c.Request.URL.String())
		fmt.Printf("\"method\": \"%s\"\n", c.Request.Method)
		fmt.Printf("\"code\": %d\n", c.Writer.Status())
		fmt.Printf("\"duration\": \"%s\"\n", duration.String())
		fmt.Printf("\"ip\": \"%s\"\n", c.ClientIP())
		fmt.Printf("\"req_headers\": %s\n", string(reqH))
		fmt.Printf("\"requestData\": %s\n", string(reqBody))
		fmt.Printf("\"resp_headers\": %s\n", string(respH))
		fmt.Printf("\"responseData\": %s\n", finalResp)
		fmt.Printf("\033[32m===========================================\033[0m\n")
	}

	// 8. 【文件 JSON 版写入】
	fields := []zap.Field{
		zap.String("url", c.Request.URL.String()),
		zap.String("method", c.Request.Method),
		zap.Int("code", c.Writer.Status()),
		zap.String("ip", c.ClientIP()),
		zap.String("duration", duration.String()),
		zap.Any("req_headers", reqHeaders),
		zap.String("requestData", string(reqBody)),
		zap.Any("resp_headers", respHeaders),
		zap.String("responseData", finalResp),
	}

	log := l.LoggerService.Log
	if err, ok := result.(error); ok {
		fields = append(fields, zap.Error(err))
		log.Error("HTTP Request Error", fields...)
	} else {
		log.Info("HTTP Request OK", fields...)
	}

	return result
}
