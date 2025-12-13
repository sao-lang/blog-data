package middlewares

import (
	"bytes"
	"io"
	"io/ioutil"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// ========================
// ResponseWriter 封装
// ========================
type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// ========================
// 初始化 zap logger
// ========================
func NewLogger(env string) (*zap.Logger, error) {
	var cfg zap.Config

	if env == "prod" {
		// 生产环境，JSON 输出，文件 + stdout
		cfg = zap.Config{
			Encoding:         "json",
			Level:            zap.NewAtomicLevelAt(zap.InfoLevel),
			OutputPaths:      []string{"stdout", "logs/app.log"},
			ErrorOutputPaths: []string{"stderr"},
			EncoderConfig: zapcore.EncoderConfig{
				TimeKey:        "time",
				LevelKey:       "level",
				NameKey:        "logger",
				CallerKey:      "caller",
				MessageKey:     "msg",
				StacktraceKey:  "stacktrace",
				LineEnding:     zapcore.DefaultLineEnding,
				EncodeLevel:    zapcore.LowercaseLevelEncoder,
				EncodeTime:     zapcore.ISO8601TimeEncoder,
				EncodeDuration: zapcore.StringDurationEncoder,
				EncodeCaller:   zapcore.ShortCallerEncoder,
			},
		}
	} else {
		// 开发环境，console 输出
		cfg = zap.Config{
			Encoding:         "console",
			Level:            zap.NewAtomicLevelAt(zap.DebugLevel),
			OutputPaths:      []string{"stdout"},
			ErrorOutputPaths: []string{"stderr"},
			EncoderConfig: zapcore.EncoderConfig{
				TimeKey:        "time",
				LevelKey:       "level",
				NameKey:        "logger",
				CallerKey:      "caller",
				MessageKey:     "msg",
				StacktraceKey:  "stacktrace",
				LineEnding:     zapcore.DefaultLineEnding,
				EncodeLevel:    zapcore.CapitalColorLevelEncoder,
				EncodeTime:     zapcore.ISO8601TimeEncoder,
				EncodeDuration: zapcore.StringDurationEncoder,
				EncodeCaller:   zapcore.ShortCallerEncoder,
			},
		}
	}

	return cfg.Build()
}

// ========================
// Gin Logger Middleware
// ========================
func Logger(env string) gin.HandlerFunc {
	logger, err := NewLogger(env)
	if err != nil {
		panic(err)
	}

	return func(c *gin.Context) {
		start := time.Now()

		// 包装 response writer
		blw := &bodyLogWriter{
			ResponseWriter: c.Writer,
			body:           &bytes.Buffer{},
		}
		c.Writer = blw

		// 读取 request body
		var reqBody []byte
		if c.Request.Body != nil {
			reqBody, _ = ioutil.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(reqBody))
		}

		// 执行 handler
		c.Next()

		// 计算耗时
		duration := time.Since(start)

		// 获取信息
		status := c.Writer.Status()
		method := c.Request.Method
		path := c.Request.URL.String()
		clientIP := c.ClientIP()
		reqStr := string(reqBody)
		respStr := blw.body.String()

		// 记录错误
		errFields := []zap.Field{}
		if len(c.Errors) > 0 {
			errFields = append(errFields, zap.String("errors", c.Errors.String()))
		}

		// 统一日志结构
		fields := []zap.Field{
			zap.Int("status", status),
			zap.String("method", method),
			zap.String("path", path),
			zap.String("ip", clientIP),
			zap.String("duration", duration.String()),
			zap.String("request", reqStr),
			zap.String("response", respStr),
		}
		fields = append(fields, errFields...)

		// 按状态码分级
		switch {
		case status >= 500:
			logger.Error("HTTP Request", fields...)
		case status >= 400:
			logger.Warn("HTTP Request", fields...)
		default:
			logger.Info("HTTP Request", fields...)
		}
	}
}
