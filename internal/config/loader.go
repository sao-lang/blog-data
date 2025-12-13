package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

type Config struct {
	SecretKey string
	ConfigKey string

	// 数据库
	PgSQL struct {
		Host     string
		Port     int
		User     string
		Password string
		DBName   string
		SSLMode  string
		MaxIdle  int
		MaxOpen  int
		LogLevel string // 或 logger.LogLevel
	}

	// Redis
	Redis struct {
		Addr     string
		Password string
		DB       int
	}

	// MinIO
	Minio struct {
		Endpoint        string
		AccessKeyID     string
		SecretAccessKey string
		UseSSL          bool
		Bucket          string
	}

	// 中间件 Key 配置
	MiddlewaresKeys struct {
		Response struct {
			Response   string
			StatusCode string
			Message    string
		}
		Validate struct {
			Validated string
		}
		Auth struct {
			TokenKey string
		}
	}
}

func LoadConfig() (*Config, error) {
	// 获取程序当前的工作目录
	currentDir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current directory: %s", err)
	}
	configFilePath := filepath.Join(currentDir, "internal", "config", "config.yaml")
	viper.SetConfigFile(configFilePath)

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}
	cfg := Config{}
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
