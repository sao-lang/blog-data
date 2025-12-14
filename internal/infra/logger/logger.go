package logger

import (
	"os"
	"path"
	"time"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type LoggerService struct {
	Log *zap.Logger
}

func NewLoggerService(env string) *LoggerService {
	logDir := "logs"
	// 确保日志目录存在
	_ = os.MkdirAll(logDir, os.ModePerm)

	// 1. 配置 rotatelogs：实现真正的日期文件名 (如 app-2025-12-14.log)
	writer, _ := rotatelogs.New(
		path.Join(logDir, "app-%Y-%m-%d.log"),
		rotatelogs.WithMaxAge(30*24*time.Hour),    // 文件保留 30 天
		rotatelogs.WithRotationTime(24*time.Hour), // 24 小时切割一次
	)

	// 2. 自定义时间格式：2025-12-14 18:00:00 (去掉毫秒)
	customTimeEncoder := func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.Format("2006-01-02 15:04:05"))
	}

	// 3. 设置 Encoder 配置
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = customTimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	// 4. 创建 MultiCore (Tee)
	// 文件输出：纯 JSON，方便 ES 分析
	// 控制台输出：Console 格式，方便开发阅读
	core := zapcore.NewTee(
		// 文件核心
		zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderConfig),
			zapcore.AddSync(writer),
			zap.InfoLevel,
		),
		// 控制台核心 (如果是开发环境，可以使用带颜色的 Encoder)
		zapcore.NewCore(
			zapcore.NewConsoleEncoder(func() zapcore.EncoderConfig {
				conf := encoderConfig
				conf.EncodeLevel = zapcore.CapitalColorLevelEncoder
				return conf
			}()),
			zapcore.AddSync(os.Stdout),
			zap.DebugLevel,
		),
	)

	// 5. 构造 Logger 实例
	// zap.AddCaller() 会记录打印日志的具体代码行号
	return &LoggerService{
		Log: zap.New(core, zap.AddCaller()),
	}
}
